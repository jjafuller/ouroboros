package command

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jjafuller/ouroboros/command/dotnet"
)

var (
	tplPath, tplName, dstPath, dstName string
	ignore                             = regexp.MustCompile(`\.vs|\.git|\.DS_Store|Thumbs.db|bin/|obj/|packages/`)
)

// DotnetCommand creates a .NET solution from a .NET solution
type DotnetCommand struct {
	Meta
}

// Run executes this command
func (c *DotnetCommand) Run(args []string) int {
	//fmt.Printf("%v", args) // dump args

	cmdFlags := flag.NewFlagSet("dotnet", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Ui.Output(c.Help()) }

	//newGuid := flag.Bool("new-guid", false, "")

	if err := cmdFlags.Parse(args); err != nil {
		c.Ui.Error(fmt.Sprintf("Invalid arguments: %s\n", err))
		return 1
	}

	if len(args) != 2 {
		c.Ui.Error("The dotnet command expects two arguments\n")
		return 1
	}

	tplPath = args[0]
	dstPath = args[1]

	tplName = path.Base(tplPath)
	dstName = path.Base(dstPath)

	fmt.Printf("Template: %s\n", tplName)
	fmt.Printf("Destination: %s\n", dstName)

	// ensure destination directory
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		os.Mkdir(dstPath, 0755)
	}

	err := filepath.Walk(tplPath, visit)
	fmt.Printf("filepath.Walk() returned %v\n", err)

	return 0
}

func visit(filePath string, fi os.FileInfo, err error) error {
	rel := strings.Replace(filePath, tplPath, "", 1)

	if len(rel) == 0 || ignore.MatchString(rel) {
		return nil
	}

	dstRel := strings.Replace(rel, tplName, dstName, -1)

	tpl := filePath
	dst := path.Join(dstPath, dstRel)

	copyErr := copy(tpl, dst, fi)
	if copyErr != nil {
		fmt.Printf("Error: %s\n", copyErr)
	}

	fmt.Printf("Visited: %s\n", rel)
	return nil
}

func copy(src, dst string, fi os.FileInfo) (err error) {
	if fi.IsDir() {
		if _, err := os.Stat(dst); os.IsNotExist(err) {
			os.Mkdir(dst, 0755)
		}
	} else {
		ext := strings.ToLower(filepath.Ext(fi.Name()))

		// if this file is a type known to contain project names,
		// edit it, otherwise just copy it
		if _, ok := dotnet.VsData.SourceFileExts[ext]; ok {
			err = copyFileAndEdit(src, dst)
		} else if _, ok := dotnet.VsData.ProjectFileExts[ext]; ok {
			err = copyFileAndEdit(src, dst)
		} else {
			err = copyFile(src, dst)
		}
	}

	return
}

func copyFileAndEdit(src, dst string) (err error) {
	read, err := ioutil.ReadFile(src)
	if err != nil {
		return
	}

	newContents := strings.Replace(string(read), tplName, dstName, -1)

	err = ioutil.WriteFile(dst, []byte(newContents), 0)

	return
}

// copyFile copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()

	return
}

// Synopsis gives an overview of this command
func (c *DotnetCommand) Synopsis() string {
	return "Create a new .NET solution from a .NET solution"
}

// Help provides detailed usage information for this command
func (c *DotnetCommand) Help() string {
	helpText := `

`
	return strings.TrimSpace(helpText)
}
