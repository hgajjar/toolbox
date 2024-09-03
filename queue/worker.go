package queue

import (
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"
	"toolbox/config"

	"github.com/gosuri/uilive"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type Worker struct {
	conn             *amqp.Connection
	queues           []string
	daemonMode       bool
	consoleCmdPrefix []string
	consoleCmdDir    string
}

type queueMessageMap map[string]int

func NewWorker(conn *amqp.Connection, queues []string, daemonMode bool) *Worker {
	return &Worker{
		conn:             conn,
		queues:           queues,
		daemonMode:       daemonMode,
		consoleCmdPrefix: strings.Split(viper.GetString(config.ConsoleCmdPrefixKey), " "),
		consoleCmdDir:    viper.GetString(config.ConsoleCmdDirKey),
	}
}

func (w *Worker) Execute(ctx context.Context) {
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				w.daemonMode = false
			}
		}
	}(ctx)

	qMap := w.initQueueMessageMap()

	queueWrite := sync.RWMutex{}
	var wg sync.WaitGroup

	writer := uilive.New()
	writer.Start()

	for queue := range qMap {
		wg.Add(1)

		go func(ctx context.Context, queue string) {
			defer wg.Done()
			w.startQueueProcess(ctx, queue, qMap, &queueWrite)
		}(ctx, queue)
	}

	go func() {
		for {
			w.printStats(qMap, writer)
			time.Sleep(time.Millisecond * 500)
		}
	}()

	wg.Wait()

	w.printStats(qMap, writer)

	writer.Stop()
}

func (w *Worker) SetDaemonMode(mode bool) {
	w.daemonMode = mode
}

func (w *Worker) printStats(queues queueMessageMap, wr io.Writer) {
	var message string
	for _, key := range w.sortMapKeys(queues) {
		message += fmt.Sprintf("Messages: %d [%s]\n", queues[key], key)
	}

	fmt.Fprint(wr, message)
}

func (w *Worker) sortMapKeys(queues queueMessageMap) []string {
	keys := make([]string, 0, len(queues))
	for k := range queues {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func (w *Worker) startQueueProcess(ctx context.Context, queue string, queues queueMessageMap, queueWrite *sync.RWMutex) {
	rmqChannel, err := w.conn.Channel()
	w.failOnError(ctx, err, "Failed to open a rabbitmq channel")
	defer rmqChannel.Close()

	for {
		q, err := rmqChannel.QueueDeclarePassive(
			queue,
			false,
			false,
			false,
			false,
			nil,
		)
		w.failOnError(ctx, err, "Failed to declare a queue")

		queueWrite.Lock()
		queues[queue] = q.Messages
		queueWrite.Unlock()

		if q.Messages > 0 {
			w.triggerQueueProcess(ctx, queue)
		} else {
			if w.daemonMode {
				time.Sleep(time.Millisecond * 100)
			} else {
				return
			}
		}
	}
}

func (w *Worker) triggerQueueProcess(ctx context.Context, queue string) {
	cmdArgs := append(w.consoleCmdPrefix, "vendor/bin/console", "queue:task:start", queue)

	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = w.consoleCmdDir

	if op, err := cmd.Output(); err != nil {
		fmt.Println(string(op))
		log.Fatal(err)
	}
}

func (w *Worker) initQueueMessageMap() queueMessageMap {
	qMap := make(queueMessageMap, len(w.queues))
	for _, queue := range w.queues {
		qMap[queue] = 0
	}

	return qMap
}

func (w *Worker) failOnError(ctx context.Context, err error, msg string) {
	if err != nil {
		zerolog.Ctx(ctx).Panic().Stack().Err(err).Msg(msg)
	}
}
