package web 

import (
	"net/http"
	"embed"
	"io/fs"
)

//go:embed static
var staticFiles embed.FS

func StaticFileHandler() http.Handler {
	sub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err)
	}
	return http.FileServer(http.FS(sub))
}
