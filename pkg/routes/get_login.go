package routes

import (
	"fmt"
	"net/http"

	"github.com/phillip-england/thispage/pkg/auth"
	"github.com/phillip-england/thispage/pkg/ratelimit"
	"github.com/phillip-england/vii/vii"
)

func GetLogin(w http.ResponseWriter, r *http.Request) {
	if auth.IsAuthenticated(r) {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	renderData := map[string]interface{}{}

	// Check rate limit status for this IP
	clientIP := ratelimit.GetClientIP(r)
	status, err := ratelimit.GetLoginStatus(clientIP)
	if err == nil {
		if status.IsBlacklisted {
			renderData["Error"] = "Your IP has been permanently blocked due to too many failed login attempts."
			renderData["IsBlocked"] = true
		} else if status.FailedAttempts > 0 && status.AttemptsLeft <= 3 {
			renderData["Warning"] = fmt.Sprintf("Warning: %d attempt(s) remaining before your IP is permanently blocked.", status.AttemptsLeft)
			renderData["AttemptsLeft"] = status.AttemptsLeft
		}
	}

	err = vii.Render(w, r, "admin_login.html", renderData)
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, err.Error())
	}
}
