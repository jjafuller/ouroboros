# Ouroboros

Ouroboros is a command line utility that can be used to recreate template solutions. Solutions for the following frameworks are supported:

* .NET

## Usage

```
usage: ouroboros [--version] [--help] <command> [<args>]

Available commands are:
    dotnet     Create a new .NET solution from a .NET solution
    version    Print ouroboros version and quit
```

### dotnet

```
usage: ouroboros dotnet <args> [tpl path] [dst path]
  
Available args are:  
    new-guids     Generate new project GUIDs
```

## Install

To install, use `go get`:

```bash
$ go get -d github.com/jjafuller/ouroboros
```

## Contribution

1. Fork ([https://github.com/jjafuller/ouroboros/fork](https://github.com/jjafuller/ouroboros/fork))
1. Create a feature branch
1. Commit your changes
1. Rebase your local changes against the master branch
1. Run test suite with the `go test ./...` command and confirm that it passes
1. Run `gofmt -s`
1. Create a new Pull Request

## Author

[jjafuller](https://github.com/jjafuller)