package recipe

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"toolbox/config"

	"github.com/spf13/viper"
)

type command struct {
	name    string
	cmd     []string
	depends []string
	groups  []string
}

func newCommand(name, cmdStr string, depends []string) command {
	return command{
		name:    name,
		cmd:     parseCmd(cmdStr),
		depends: depends,
	}
}

func parseCmd(cmdStr string) []string {
	cmdPrefix := strings.FieldsFunc(viper.GetString(config.ConsoleCmdPrefixKey), func(c rune) bool {
		return c == ' '
	})

	cmd := cmdPrefix

	r := csv.NewReader(strings.NewReader(cmdStr))
	r.Comma = ' '
	r.LazyQuotes = true
	for {
		cmdPart, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		cmd = append(cmd, cmdPart...)
	}

	return cmd
}

func (c *command) Trigger(ctx context.Context) ([]byte, error) {
	cmdDir := viper.GetString(config.ConsoleCmdDirKey)
	// commandParts := append(i.consoleCmdPrefix, strings.Split(command, " ")...)

	cmd := exec.CommandContext(ctx, c.cmd[0], c.cmd[1:]...)
	cmd.Dir = cmdDir

	op, err := cmd.Output()
	if err != nil {
		return op, err
	}

	if exitCode := cmd.ProcessState.ExitCode(); exitCode != 0 {
		return op, fmt.Errorf("process returned a non-zero exit code: %d", exitCode)
	}

	return op, nil
}
