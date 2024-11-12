package main

import (
	_ "embed"
	"toolbox/application"
)

var (
	//go:embed toolbox.yml
	defaultConfig []byte

	toolbox *application.Toolbox
)

func init() {
	toolbox = application.New(defaultConfig)
}

func main() {
	toolbox.Execute()
}
