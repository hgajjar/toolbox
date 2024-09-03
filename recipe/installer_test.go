package recipe

import (
	"context"
	"os"
	"testing"

	"github.com/go-test/deep"
	"github.com/rs/zerolog"
)

type logHook struct {
	logEvents []zerolog.Event
	messages  []string
}

func (logHook *logHook) Run(logEvent *zerolog.Event, level zerolog.Level, message string) {
	logHook.logEvents = append(logHook.logEvents, *logEvent)
	logHook.messages = append(logHook.messages, message)
}

func TestInstaller_Install(t *testing.T) {
	installer := Installer{
		recipe: &recipe{
			sections: []section{
				{
					name: "A",
					commands: commands{
						parallel: true,
						commands: []command{
							{
								name: "AA",
								cmd:  []string{"echo", "AA"},
							},
							{
								name:    "BB",
								cmd:     []string{"echo", "BB"},
								depends: []string{"AA"},
							},
							{
								name:    "CC",
								cmd:     []string{"echo", "CC"},
								depends: []string{"BB"},
							},
							{
								name:    "DD",
								cmd:     []string{"echo", "DD"},
								depends: []string{"BB", "FF"},
							},
							{
								name:    "EE",
								cmd:     []string{"echo", "EE"},
								depends: []string{"DD"},
							},
							{
								name:    "FF",
								cmd:     []string{"echo", "FF"},
								depends: []string{"CC"},
							},
							{
								name:    "ZZ",
								cmd:     []string{"echo", "ZZ"},
								depends: []string{"*"},
							},
						},
					},
				},
				{
					name: "B",
					commands: commands{
						commands: []command{
							{
								name: "BBB",
								cmd:  []string{"echo", "BBB"},
							},
						},
					},
				},
				{
					name:  "C",
					async: true,
					commands: commands{
						parallel: true,
						commands: []command{
							{
								name: "CCC",
								cmd:  []string{"echo", "CCC"},
							},
						},
					},
				},
			},
		},
	}

	logger := zerolog.New(os.Stdout)
	logHook := &logHook{}
	logger = logger.Hook(logHook)

	// Attach the Logger to the context.Context
	ctx := logger.WithContext(context.Background())

	installer.Install(ctx)

	expected := []string{
		"Section: A\n\n",
		"Command: AA",
		"AA\n",
		"Command: BB",
		"BB\n",
		"Command: CC",
		"CC\n",
		"Command: FF",
		"FF\n",
		"Command: DD",
		"DD\n",
		"Command: EE",
		"EE\n",
		"Command: ZZ",
		"ZZ\n",
		"Section: B\n\n",
		"Command: BBB",
		"BBB\n",
		"Section: C\n\n",
		"Command: CCC\nCCC\n",
	}

	if diff := deep.Equal(expected, logHook.messages); diff != nil {
		t.Error(diff)
	}
}
