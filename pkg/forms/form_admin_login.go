package forms

import (
	"net/http"
)

type FormAdminLoginData struct {
	Username string
	Password string
}

type FormAdminLogin struct{}

func (FormAdminLogin) Validate(r *http.Request) (FormAdminLoginData, error) {
	if err := r.ParseForm(); err != nil {
		return FormAdminLoginData{}, err
	}
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	return FormAdminLoginData{
		Username: username,
		Password: password,
	}, nil
}
