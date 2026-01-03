package routes

import (
	"fmt"
	"net/http"

	"github.com/phillip-england/thispage/pkg/auth"
	"github.com/phillip-england/thispage/pkg/credentials"
	"github.com/phillip-england/thispage/pkg/database"
	"github.com/phillip-england/thispage/pkg/forms"
	"github.com/phillip-england/thispage/pkg/keys"
	"github.com/phillip-england/thispage/pkg/ratelimit"
	"github.com/phillip-england/vii/vii"
)

func PostLogin(w http.ResponseWriter, r *http.Request) {
	clientIP := ratelimit.GetClientIP(r)

	// Check if IP is blacklisted
	status, err := ratelimit.GetLoginStatus(clientIP)
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, "Failed to check rate limit: "+err.Error())
		return
	}

	if status.IsBlacklisted {
		vii.Render(w, r, "admin_login.html", map[string]interface{}{
			"Error":       "Your IP has been permanently blocked due to too many failed login attempts.",
			"IsBlocked":   true,
		})
		return
	}

	validator := forms.FormAdminLogin{}
	data, err := validator.Validate(r)
	if err != nil {
		vii.Render(w, r, "admin_login.html", map[string]interface{}{"Error": "Invalid form data"})
		return
	}

	projectPath, ok := vii.GetContext(keys.ProjectPath, r).(string)
	if !ok {
		vii.WriteError(w, http.StatusInternalServerError, "Project path not found in context")
		return
	}

	adminUsername, adminPassword, err := credentials.Load(projectPath)
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, "Failed to load credentials: "+err.Error())
		return
	}

	if data.Username != adminUsername || data.Password != adminPassword {
		// Record failed attempt
		shouldBlacklist, err := ratelimit.RecordAttempt(clientIP, false)
		if err != nil {
			vii.WriteError(w, http.StatusInternalServerError, "Failed to record login attempt: "+err.Error())
			return
		}

		if shouldBlacklist {
			// Add to blacklist
			if err := ratelimit.AddToBlacklist(clientIP); err != nil {
				vii.WriteError(w, http.StatusInternalServerError, "Failed to update blacklist: "+err.Error())
				return
			}
			vii.Render(w, r, "admin_login.html", map[string]interface{}{
				"Error":     "Too many failed login attempts. Your IP has been permanently blocked.",
				"IsBlocked": true,
			})
			return
		}

		// Get updated status for warning
		newStatus, _ := ratelimit.GetLoginStatus(clientIP)
		attemptsLeft := newStatus.AttemptsLeft

		renderData := map[string]interface{}{
			"Error": "Invalid credentials",
		}

		// Show warning if they're getting close to being locked out
		if attemptsLeft <= 3 && attemptsLeft > 0 {
			renderData["Warning"] = fmt.Sprintf("Warning: %d attempt(s) remaining before your IP is permanently blocked.", attemptsLeft)
			renderData["AttemptsLeft"] = attemptsLeft
		} else if attemptsLeft == 0 {
			renderData["Warning"] = "This is your final attempt. You will be locked out after this."
			renderData["AttemptsLeft"] = 0
		}

		vii.Render(w, r, "admin_login.html", renderData)
		return
	}

	// Successful login - record it and clear previous failed attempts
	ratelimit.RecordAttempt(clientIP, true)
	ratelimit.ClearAttemptsForIP(clientIP)

	// Create Session
	if err := auth.CreateSession(w, r); err != nil {
		vii.WriteError(w, http.StatusInternalServerError, "Failed to create session: "+err.Error())
		return
	}

	// Success
	vii.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// CheckLoginRateLimit is a helper to check rate limit status (for GET /login)
func CheckLoginRateLimit(r *http.Request) (status ratelimit.LoginStatus, err error) {
	clientIP := ratelimit.GetClientIP(r)
	return ratelimit.GetLoginStatus(clientIP)
}

// Ensure database constants are accessible for display
var _ = database.FailedAttemptThreshold
