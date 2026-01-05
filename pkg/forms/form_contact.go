package forms

import (
	"errors"
	"net/http"
	"strings"
)

const (
	MaxNameLength    = 100
	MaxEmailLength   = 255
	MaxMessageLength = 2000
)

type FormContactData struct {
	Name    string
	Email   string
	Message string
}

type FormContact struct{}

func (FormContact) Validate(r *http.Request) (FormContactData, error) {
	if err := r.ParseForm(); err != nil {
		return FormContactData{}, err
	}

	name := strings.TrimSpace(r.Form.Get("name"))
	email := strings.TrimSpace(r.Form.Get("email"))
	message := strings.TrimSpace(r.Form.Get("message"))

	// Validate required fields
	if name == "" {
		return FormContactData{}, errors.New("name is required")
	}
	if email == "" {
		return FormContactData{}, errors.New("email is required")
	}
	if message == "" {
		return FormContactData{}, errors.New("message is required")
	}

	// Validate lengths
	if len(name) > MaxNameLength {
		return FormContactData{}, errors.New("name must be 100 characters or less")
	}
	if len(email) > MaxEmailLength {
		return FormContactData{}, errors.New("email must be 255 characters or less")
	}
	if len(message) > MaxMessageLength {
		return FormContactData{}, errors.New("message must be 2000 characters or less")
	}

	// Basic email validation
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return FormContactData{}, errors.New("please enter a valid email address")
	}

	return FormContactData{
		Name:    name,
		Email:   email,
		Message: message,
	}, nil
}
