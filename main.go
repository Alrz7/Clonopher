package main

import (
	"clonopher/app"
	"clonopher/internal/loggy"
)

func main() {
	if err := app.Main(); err != nil {
		loggy.Fatal(err)
	}
}
