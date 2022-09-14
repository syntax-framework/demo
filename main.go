package main

import (
	"github.com/syntax-framework/demo/web/controllers"
	"github.com/syntax-framework/syntax/syntax"
	"log"
	"net/http"
	"os"
)

func main() {
	handler := createSite()
	httpAddr := "localhost:8080"
	if err := http.ListenAndServeTLS(httpAddr, "localhost.crt", "localhost.key", handler); err != nil {
		log.Fatalf("ListenAndServe %s: %v", httpAddr, err)
	}
}

func createSite() http.Handler {

	//syntax.LoadConfig()

	app := syntax.New(nil)

	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	app.AddFileSystemDir(path+"/web/", 0)

	//site.Midleware()

	// go:embed site_embed/*
	//var embedSiteDir embed.FS
	//site.AddFileSystemEmbed(embedSiteDir, "site_embed/", 0) // test only

	controllers.RegisterMyController(app)
	controllers.RegisterMyLiveController(app)

	if err := app.Init(); err != nil {
		log.Fatal(err)
	}

	return app.Handler
}
