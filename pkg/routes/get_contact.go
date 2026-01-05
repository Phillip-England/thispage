package routes

import (
	"net/http"

	"github.com/phillip-england/vii/vii"
)

func GetContact(w http.ResponseWriter, r *http.Request) {
	remaining, err := GetMessageStatus(r)
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, "Failed to check message limit: "+err.Error())
		return
	}

	data := map[string]interface{}{
		"MessagesRemaining": remaining,
	}

	if remaining <= 0 {
		data["IsBlocked"] = true
	}

	vii.Render(w, r, "admin_contact.html", data)
}
