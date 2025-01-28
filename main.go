package main

import (
	_ "embed"

	"github.com/hgajjar/toolbox/application"
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
