package core

import (
	"clonopher/app/burrow"
	"clonopher/app/jevkit"
	"clonopher/app/services/composer"
	"testing"
)

func TestMain(t *testing.T) {
	Payload := Payload{
		Name: "effective-go.pdf",
		Head: "/home/alireza/Documents",
		Src:  "/home/alireza/Documents/go-books/effective-go.pdf",
		Info: &composer.FileInfo{
			Description: "A reference book covering idiomatic Go development, effective language patterns, code organization, concurrency practices, error handling, and common approaches for writing clear and maintainable Go programs.",
		},
		Stage: &burrow.Stage{Path: "/home/alireza/Documents/alrzvalut-obsidian",
			Roots: map[string]*burrow.Stage{}, Info: &composer.StageInfo{Description: ""}},
		Options: &Options{
			Provider: &jevkit.Provider{Name: "ollama",
				Url:    "http://127.0.0.1:11434/v1/systemone",
				Models: []string{"tev1:0.8b"}},
			Model: "tev1:0.8b",
		},
	}
	err := Payload.Serve()
	if err != nil {
		t.Log(err)
	}
}
