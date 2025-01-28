//go:build integration
// +build integration

package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/hgajjar/toolbox/integration/rabbitmq"
)

func main() {
	args := os.Args[1:]

	if len(args) != 3 {
		log.Fatalf("expected 3 arguments, got %d. Usage: ./consumer [rabbutmq-port] [chunk-size] [queue-name]", len(args))
	}

	rmq, err := rabbitmq.New(rabbitmq.Config{
		URL: fmt.Sprintf("amqp://guest:guest@127.0.0.1:%s/", args[0]),
	})
	if err != nil {
		log.Panic(err)
	}
	defer rmq.Close()

	size, err := strconv.Atoi(args[1])
	if err != nil {
		log.Panic(err)
	}

	mCh, err := rmq.Consume(args[2], size)
	if err != nil {
		log.Panic(err)
	}

	for m := range mCh {
		fmt.Println("Message received: ", string(m))
	}
}
