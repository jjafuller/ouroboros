package command

import (
	"testing"

	"github.com/mitchellh/cli"
)

func TestDotnetCommand_implement(t *testing.T) {
	var _ cli.Command = &DotnetCommand{}
}

func TestDotnetCommand_extractGUIDSln(t *testing.T) {
	var command = &DotnetCommand{}

	contents := `
Project("{FAE04EC0-301F-11D3-BF4B-00C04F79EFBC}") = "Test", "Test\Test.csproj", "{8EA60CA5-7D3D-4813-ACB1-069618285452}"
EndProject
Project("{FAE04EC0-301F-11D3-BF4B-00C04F79EFBC}") = "Test.Tests", "Test.Tests\Test.Tests.csproj", "{E20EF42E-DBB7-42AF-B870-6B1C9D27FC48}"
EndProject
`

	command.ExtractGUID(contents, ".sln")
}

func TestDotnetCommand_extractGUIDProj(t *testing.T) {
	var command = &DotnetCommand{}

	contents := `
    <Platform Condition=" '$(Platform)' == '' ">AnyCPU</Platform>
    <ProjectGuid>{E20EF42E-DBB7-42AF-B870-6B1C9D27FC48}</ProjectGuid>
    <OutputType>Library</OutputType>
`

	command.ExtractGUID(contents, "csproj")
}
