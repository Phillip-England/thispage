package routes

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/phillip-england/thispage/pkg/auth"
	"github.com/phillip-england/vii/vii"
)

func GetAdminLogout(w http.ResponseWriter, r *http.Request) {
	_ = auth.DeleteSession(w, r)

	nextPath := sanitizeNextPath(r.URL.Query().Get("next"))
	if nextPath == "" {
		nextPath = "/"
	}
	vii.Redirect(w, r, nextPath, http.StatusSeeOther)
}

func sanitizeNextPath(next string) string {
	if next == "" {
		return ""
	}

	parsed, err := url.Parse(next)
	if err != nil {
		return ""
	}

	if parsed.Scheme != "" || parsed.Host != "" {
		return ""
	}

	if !strings.HasPrefix(parsed.Path, "/") || strings.HasPrefix(parsed.Path, "//") {
		return ""
	}

	query := parsed.Query()
	query.Del("is_admin")
	parsed.RawQuery = query.Encode()

	if parsed.Path == "" {
		parsed.Path = "/"
	}

	return parsed.RequestURI()
}
