package main

import (
	"os"
	"testing"
	"toolbox/application"
)

func TestApp(t *testing.T) {
	os.Args = append(os.Args, "sync:data", "-r", "product_abstract")

	toolbox := application.New()

	toolbox.Execute()
}
