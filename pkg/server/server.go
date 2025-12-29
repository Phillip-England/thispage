package server

import (
	"fmt"
	"net/http"
)

func Serve(dir string, port string) error {
	fs := http.FileServer(http.Dir(dir))
	http.Handle("/", fs)

	fmt.Printf("Serving %s on http://localhost:%s\n", dir, port)
	return http.ListenAndServe(":"+port, nil)
}
