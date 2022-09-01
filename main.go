package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/syntax-framework/syntax/syntax"
	"log"
	"net/http"
	"os"
)

func main() {
	handler := createSite()
	httpAddr := "localhost:8080"
	if err := http.ListenAndServe(httpAddr, handler); err != nil {
		log.Fatalf("ListenAndServe %s: %v", httpAddr, err)
	}
}

func createSite() *httprouter.Router {

	config := &syntax.Config{
		Dev: true,
		LiveReload: syntax.ConfigLiveReload{
			Interval: 100,
			Debounce: 200,
			//ReloadPageOnCss: false,
			//Patterns:        nil,
			//Endpoint:        "",
		},
	}

	//syntax.LoadConfig()

	site := syntax.New(nil)

	if config.Dev {
		path, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		site.AddFileSystemDir(path+"/web/", 0)
	}

	// go:embed site_embed/*
	//var embedSiteDir embed.FS
	//site.AddFileSystemEmbed(embedSiteDir, "site_embed/", 0) // test only

	if err := site.Init(); err != nil {
		log.Fatal(err)
	}

	return site.Router
}
