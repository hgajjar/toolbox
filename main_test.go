package main

import (
	"os"
	"queue-worker/application"
	"testing"
)

func TestApp(t *testing.T) {
	os.Args = append(os.Args, "sync:data", "-r", "product_abstract")

	toolbox := application.New()

	toolbox.Execute()
}
