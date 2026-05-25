package ui

import (
	"embed"
	"net/http"
)

//go:embed static/*
var staticFiles embed.FS

func Handler() http.Handler {
	files := http.FileServer(http.FS(staticFiles))
	mux := http.NewServeMux()
	mux.Handle("/static/", files)
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/" {
			http.NotFound(writer, request)
			return
		}
		http.ServeFileFS(writer, request, staticFiles, "static/index.html")
	})
	return mux
}
