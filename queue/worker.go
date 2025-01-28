package queue

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"sort"
	"sync"
	"time"

	"github.com/hgajjar/toolbox/config"

	"github.com/Adaendra/uilive"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

type Worker struct {
	conn             *amqp.Connection
	queues           []string
	daemonMode       bool
	consoleCmdPrefix []string
	consoleCmdDir    string
	consoleCmd       []string
}

type queueMessageMap map[string]int

func NewWorker(conn *amqp.Connection, queues []string, daemonMode bool, cmdPrefix []string, cmdDir string, cmd []string) *Worker {
	return &Worker{
		conn:             conn,
		queues:           queues,
		daemonMode:       daemonMode,
		consoleCmdPrefix: cmdPrefix,
		consoleCmdDir:    cmdDir,
		consoleCmd:       cmd,
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

	queueMapLock := sync.RWMutex{}
	var wg sync.WaitGroup

	writer := uilive.New()
	writer.Start()

	for queue := range qMap {
		wg.Add(1)

		go func(ctx context.Context, queue string) {
			defer wg.Done()
			err := w.startQueueProcess(ctx, queue, qMap, &queueMapLock)
			if err != nil {
				zerolog.Ctx(ctx).Error().Stack().Err(err).Msg(err.Error())
			}
		}(ctx, queue)
	}

	if config.Verbose {
		go func(qMap queueMessageMap, writer io.Writer, queueMapLock *sync.RWMutex) {
			for {
				w.printStats(qMap, writer, queueMapLock)
				time.Sleep(time.Millisecond * 500)
			}
		}(qMap, writer, &queueMapLock)
	}

	wg.Wait()

	if config.Verbose {
		w.printStats(qMap, writer, &queueMapLock)
	}

	writer.Stop()
}

func (w *Worker) SetDaemonMode(mode bool) {
	w.daemonMode = mode
}

func (w *Worker) printStats(queues queueMessageMap, wr io.Writer, queueMapLock *sync.RWMutex) {
	queueMapLock.RLock()
	defer queueMapLock.RUnlock()

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

func (w *Worker) startQueueProcess(ctx context.Context, queue string, queues queueMessageMap, queueMapLock *sync.RWMutex) error {
	rmqChannel, err := w.conn.Channel()
	if err != nil {
		return errors.Wrap(err, "Failed to open a rabbitmq channel")
	}
	defer rmqChannel.Close()

	var i int

	for {
		q, err := rmqChannel.QueueDeclarePassive(
			queue,
			false,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return errors.Wrap(err, "Failed to declare a queue")
		}

		queueMapLock.Lock()
		queues[queue] = q.Messages
		queueMapLock.Unlock()

		if q.Messages > 0 {
			err := w.triggerQueueProcess(ctx, queue)
			if err != nil {
				return err
			}
		} else {
			if w.daemonMode {
				time.Sleep(time.Millisecond * 100)
			} else {
				if i == 0 {
					// wait once if there are any messages added in queue after it's started
					i++
					time.Sleep(time.Second * 1)
				}

				if w.areAllQueuesEmpty(queues, queueMapLock) {
					return nil
				}

				time.Sleep(time.Millisecond * 100)
			}
		}
	}
}

func (w *Worker) areAllQueuesEmpty(queues queueMessageMap, queueMapLock *sync.RWMutex) bool {
	queueMapLock.RLock()
	defer queueMapLock.RUnlock()

	for _, messageCount := range queues {
		if messageCount > 0 {
			return false
		}
	}

	return true
}

func (w *Worker) triggerQueueProcess(ctx context.Context, queue string) error {
	var cmdArgs []string
	if len(w.consoleCmdPrefix) > 0 && w.consoleCmdPrefix[0] != "" {
		cmdArgs = append(cmdArgs, w.consoleCmdPrefix...)
	}

	cmdArgs = append(cmdArgs, w.consoleCmd...)
	cmdArgs = append(cmdArgs, queue)

	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = w.consoleCmdDir

	if op, err := cmd.Output(); err != nil {
		return errors.Wrap(err, "Failed to execute command: "+string(op))
	}

	return nil
}

func (w *Worker) initQueueMessageMap() queueMessageMap {
	qMap := make(queueMessageMap, len(w.queues))
	for _, queue := range w.queues {
		qMap[queue] = 0
	}

	return qMap
}
