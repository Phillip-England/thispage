package validators

import (
	"net/http"
)

type FormAdminLoginData struct {
	Password string
}

type FormAdminLogin struct{}

func (FormAdminLogin) Validate(r *http.Request) (FormAdminLoginData, error) {
	if err := r.ParseForm(); err != nil {
		return FormAdminLoginData{}, err
	}
	password := r.Form.Get("password")
	return FormAdminLoginData{
		Password: password,
	}, nil
}
