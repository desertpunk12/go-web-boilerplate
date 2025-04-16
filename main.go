package main

import (
	"fmt"
	"go-web-boilerplate/assets"
	"go-web-boilerplate/internal/hr/config"
	"go-web-boilerplate/ui/pages"
	"net/http"

	"github.com/a-h/templ"
)

func main() {
	mux := http.NewServeMux()
	SetupAssetsRoutes(mux)
	mux.Handle("GET /", templ.Handler(pages.Login()))
	fmt.Printf("Server is running on http://localhost:%s\n", config.PORT)
	err := http.ListenAndServe(fmt.Sprintf("%s:%s", "localhost", config.PORT), mux)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}

}

func SetupAssetsRoutes(mux *http.ServeMux) {
	assetHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if config.IS_PROD {
			w.Header().Set("Cache-Control", "no-store")
		}

		var fs http.Handler
		if config.IS_PROD {
			fs = http.FileServer(http.Dir("./assets"))
		} else {
			fs = http.FileServer(http.FS(assets.Assets))
		}

		fs.ServeHTTP(w, r)
	})

	mux.Handle("GET /assets/", http.StripPrefix("/assets/", assetHandler))
}
