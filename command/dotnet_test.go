package command

import (
	"testing"

	"github.com/mitchellh/cli"
)

func TestDotnetCommand_implement(t *testing.T) {
	var _ cli.Command = &DotnetCommand{}
}
