module github.com/phillip-england/thispage

go 1.25.5

require (
	github.com/fsnotify/fsnotify v1.9.0
	github.com/joho/godotenv v1.5.1
	github.com/phillip-england/vii v0.0.17
	github.com/spf13/cobra v1.10.2
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	golang.org/x/sys v0.39.0 // indirect
)

replace github.com/phillip-england/vii => ./pkg/vii
