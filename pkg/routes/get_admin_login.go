package routes

import (
	"net/http"

	"github.com/phillip-england/vii/vii"
)

type GetAdminLogin struct{}

func (GetAdminLogin) OnMount(app *vii.App) error {
	return nil
}

func (GetAdminLogin) Services() []vii.Service {
	return []vii.Service{}
}

func (GetAdminLogin) Validators() []vii.AnyValidator {
	return []vii.AnyValidator{}
}

func (GetAdminLogin) Handle(r *http.Request, w http.ResponseWriter) error {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	vii.Render(r, w, "templates", "admin_login.html", nil, nil)
	return nil
}

func (GetAdminLogin) OnErr(r *http.Request, w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
