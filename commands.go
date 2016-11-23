package main

import (
	"github.com/jjafuller/ouroboros/command"
	"github.com/mitchellh/cli"
)

// Commands returns the list of runnable sub-commands
func Commands(meta *command.Meta) map[string]cli.CommandFactory {
	return map[string]cli.CommandFactory{
		"dotnet": func() (cli.Command, error) {
			return &command.DotnetCommand{
				Meta: *meta,
			}, nil
		},

		"version": func() (cli.Command, error) {
			return &command.VersionCommand{
				Meta:     *meta,
				Version:  Version,
				Revision: GitCommit,
				Name:     Name,
			}, nil
		},
	}
}
