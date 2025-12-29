package routes

import (
	"io"
	"net/http"

	"github.com/phillip-england/vii/vii"
)

// home route
type BunProxy struct{}

func (BunProxy) OnMount(app *vii.App) error {
	return nil
}

func (BunProxy) Handle(r *http.Request, w http.ResponseWriter) error {
	// Build target URL
	targetURL := "http://localhost:3030" + r.URL.RequestURI()

	// Create outbound request
	req, err := http.NewRequest(
		r.Method,
		targetURL,
		r.Body,
	)
	if err != nil {
		return err
	}

	// Copy headers
	req.Header = r.Header.Clone()

	// Optional but recommended
	req.Host = "localhost:3030"

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Copy response headers
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	// Write status code
	w.WriteHeader(resp.StatusCode)

	// Stream body
	_, err = io.Copy(w, resp.Body)
	return err
}

func (BunProxy) OnErr(r *http.Request, w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusBadGateway)
}
