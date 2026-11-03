package core

import (
	"clonopher/internal/fileSystem"
	"path/filepath"
	"testing"
)

// func TestServePayload(t *testing.T) {
// 	ollamaTev := &Options{
// 		Provider: &providers.Provider{
// 			Url:    "http://127.0.0.1:11434/v1/systemone",
// 			Models: []string{"tev1:0.8b"}, ActiveModel: "tev1:0.8b"},
// 	}

// 	Payload := Payload{
// 		Name: "golangGinRateLimiter.md",
// 		Head: "/home/alireza/GoPrj/Clonopher/tmp/testing",
// 		Src:  "/home/alireza/GoPrj/Clonopher/tmp/golangGinRateLimiter.md",
// 		Info: &classifiers.ObjectInfo{
// 			Description: "A Go-based rate-limiting middleware for Gin APIs, covering request throttling, HTTP 429 responses, and per-route rate limits.",
// 		},
// 		Directory:   burrow.New("/home/alireza/GoPrj/Clonopher/tmp/testing"),
// 		Options: ollamaTev,
// 	}
// 	err := Payload.Serve()
// 	if err != nil {
// 		t.Log(err)
// 	}
// }

func TestLoadExistingClss(t *testing.T) {
	cnf := ClsConfig{Classifiers: map[string]*CustomCls{}}
	appdir, err := fileSystem.SetAppDir()
	if err != nil {
		t.Fatal(err)
	}
	err = cnf.LoadExistingClassifiers(filepath.Join(appdir, fileSystem.ClsDirName))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(cnf.Classifiers)
}
