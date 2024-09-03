package recipe

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

type Installer struct {
	recipe *recipe
}

func NewInstaller(recipeFile string) *Installer {
	return &Installer{
		recipe: newRecipe(recipeFile),
	}
}

func (i *Installer) Install(ctx context.Context) {
	wg, outputLogCh := i.triggerAsyncSections(ctx)

	i.triggerSections(ctx)

	wg.Wait()
	close(outputLogCh)

	for log := range outputLogCh {
		zerolog.Ctx(ctx).Info().Msg(log)
	}
}

func (i *Installer) triggerAsyncSections(ctx context.Context) (*sync.WaitGroup, chan string) {
	var wg sync.WaitGroup

	asyncLogCh := make(chan string, 10000)
	logWrite := sync.RWMutex{}

	for _, section := range i.recipe.sections {
		if !section.async {
			continue
		}

		wg.Add(1)

		go func() {
			defer wg.Done()

			var outputLog []string

			if section.commands.parallel {
				var err error
				outputLog, err = i.triggerAsyncCommands(ctx, section.commands)
				failOnError(ctx, err, strings.Join(outputLog, "\n"))
			} else {
				for _, command := range section.commands.commands {
					outputLog = append(outputLog, fmt.Sprintf("Command: %s", command.name))

					output, err := command.Trigger(ctx)
					failOnError(ctx, err, string(output))
					outputLog = append(outputLog, string(output))
				}
			}

			logWrite.Lock()
			asyncLogCh <- fmt.Sprintf("Section: %s\n\n", section.name)
			asyncLogCh <- strings.Join(outputLog, "\n")
			logWrite.Unlock()

		}()
	}

	return &wg, asyncLogCh
}

func (i *Installer) triggerSections(ctx context.Context) {
	for _, section := range i.recipe.sections {
		if section.async {
			continue
		}

		zerolog.Ctx(ctx).Info().Msg(fmt.Sprintf("Section: %s\n\n", section.name))

		if section.commands.parallel {
			outputLog, err := i.triggerAsyncCommands(ctx, section.commands)
			for _, log := range outputLog {
				zerolog.Ctx(ctx).Info().Msg(log)
			}
			failOnError(ctx, err, "")
		} else {
			for _, command := range section.commands.commands {
				zerolog.Ctx(ctx).Info().Msg(fmt.Sprintf("Command: %s", command.name))

				output, err := command.Trigger(ctx)
				zerolog.Ctx(ctx).Info().Msg(string(output))
				failOnError(ctx, err, "")
			}
		}
	}
}

func (i *Installer) triggerAsyncCommands(ctx context.Context, commands commands) ([]string, error) {
	errs, ctx := errgroup.WithContext(ctx)

	var asyncLog []string
	logWrite := sync.RWMutex{}

	dependencyMap := make(map[string]chan any)
	dependencyMapLock := sync.RWMutex{}

	for _, command := range commands.commands {
		doneCh := make(chan any)

		dependencyMapLock.Lock()
		dependencyMap[command.name] = doneCh
		dependencyMapLock.Unlock()

		errs.Go(func() error {
			defer func() {
				dependencyMapLock.RLock()
				close(dependencyMap[command.name])
				dependencyMapLock.RUnlock()
			}()

			waitChannels, err := i.findDependingChannels(command, commands, dependencyMap, &dependencyMapLock)
			if err != nil {
				return err
			}

			for _, waitCh := range waitChannels {
				<-waitCh
			}

			output, err := command.Trigger(ctx)

			logWrite.Lock()

			asyncLog = append(asyncLog, fmt.Sprintf("Command: %s", command.name))
			asyncLog = append(asyncLog, string(output))
			logWrite.Unlock()

			return err
		})
	}

	return asyncLog, errs.Wait()
}

func (i *Installer) findDependingChannels(command command, allCommands commands, dependencyMap map[string]chan any, lock *sync.RWMutex) ([]<-chan any, error) {
	var dependingChannels []<-chan any
	for _, depend := range command.depends {
		for _, c := range allCommands.commands {
			if depend == "*" || c.name == depend {
				if command.name == c.name {
					// a command can not depend on itself
					continue
				}

				lock.RLock()
				dependingCh, ok := dependencyMap[c.name]
				lock.RUnlock()

				if !ok {
					return nil, fmt.Errorf("command '%s' depends on '%s', but it could not find the wait channel for it", command.name, depend)
				}
				dependingChannels = append(dependingChannels, dependingCh)
			}
		}
	}

	return dependingChannels, nil
}

func failOnError(ctx context.Context, err error, msg string) {
	if err != nil {
		var exitErr *exec.ExitError

		if errors.As(err, &exitErr) {
			msg += string(exitErr.Stderr)
		}
		zerolog.Ctx(ctx).Panic().Stack().Err(err).Msg(msg)
	}
}
