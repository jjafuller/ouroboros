package main

// Name is the name of the binary
const Name string = "ouroboros"

// Version is the current version of this release
const Version string = "0.2.0"

// GitCommit describes latest commit hash.
// This value is extracted by git command when building.
// To set this from outside, use go build -ldflags "-X main.GitCommit \"$(COMMIT)\""
var GitCommit string
