package main

import (
	"queue-worker/application"
)

var toolbox *application.Toolbox

func init() {
	toolbox = application.New()
}

func main() {
	toolbox.Execute()
}
