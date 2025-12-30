package routes

import (
	"net/http"

	"github.com/phillip-england/vii/vii"
)

type PostAdminLogin struct{}

func (PostAdminLogin) OnMount(app *vii.App) error {
	return nil
}

func (PostAdminLogin) Services() []vii.Service {
	return []vii.Service{}
}

func (PostAdminLogin) Validators() []vii.AnyValidator {
	return []vii.AnyValidator{
	}
}

func (PostAdminLogin) Handle(r *http.Request, w http.ResponseWriter) error {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	vii.Render(r, w, "templates", "admin_login.html", nil, nil)
	return nil
}

func (PostAdminLogin) OnErr(r *http.Request, w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
