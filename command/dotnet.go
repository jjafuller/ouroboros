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
	"sort"
	"strings"

	"github.com/jjafuller/ouroboros/command/dotnet"
	"github.com/satori/go.uuid"
)

// in solution
// ^Project.*{(.*)}"$
// Project("{FAE04EC0-301F-11D3-BF4B-00C04F79EFBC}") = "Test", "Test\Test.csproj", "{8EA60CA5-7D3D-4813-ACB1-069618285452}"

// in project
// ^\s*<ProjectGuid>{(.*)}</ProjectGuid>\s*$
//     <ProjectGuid>{8EA60CA5-7D3D-4813-ACB1-069618285452}</ProjectGuid>

var (
	tplPath, tplName, dstPath, dstName string
	ignoreFiles                        = regexp.MustCompile(`\.DS_Store|Thumbs.db`)
	ignoreDirectories                  = regexp.MustCompile(`\.vs|\.git|bin|obj|packages`)
	guidInSln                          = regexp.MustCompile(`^Project.*{(?P<guid>.*)}"$`)
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

	newGUIDs := cmdFlags.Bool("new-guids", false, "generate new guids?")

	overrideTplName := cmdFlags.String("tpl-name", "", "override template solution name")

	if err := cmdFlags.Parse(args); err != nil {
		c.Ui.Error(fmt.Sprintf("Invalid arguments: %s\n", err))
		return 1
	}

	if len(cmdFlags.Args()) != 2 {
		c.Ui.Error("The dotnet command expects two arguments\n")
		return 1
	}

	tplPath = cmdFlags.Arg(0)
	dstPath = cmdFlags.Arg(1)

	if len(*overrideTplName) < 1 {
		tplName = filepath.Base(tplPath)
	} else {
		tplName = *overrideTplName
	}

	dstName = filepath.Base(dstPath)

	fmt.Printf("Template: \n  %s\n", tplName)
	fmt.Printf("Destination: \n  %s\n", dstName)

	// ensure destination directory
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		os.Mkdir(dstPath, 0755)
	}

	dirList, slnPath, err := getDirectoryList(tplPath)

	if err != nil {
		fmt.Printf("getDirectoryList() returned %v\n", err)
		return 1
	}

	if len(slnPath) < 1 {
		fmt.Println("No solution file was found within the directory")
		return 1
	}

	fmt.Printf("Solution file: \n  %s\n", slnPath)

	var guids map[string]string

	if *newGUIDs {
		fmt.Println("Generating new guids...")

		guidList, err := c.extractGUIDsFromSln(slnPath)

		if err != nil {
			fmt.Printf("Failed to extract guids from the solution file: %v\n", err)
			return 1
		}

		guids = c.GeneratenewGUIDs(guidList)

		for k, v := range guids {
			fmt.Printf("  %s -> %s\n", k, v)
		}
	}

	err = copyItems(dirList, guids)

	if err != nil {
		fmt.Printf("Failed generate new items: %v\n", err)
		return 1
	}

	return 0
}

func getDirectoryList(dirPath string) (map[string]os.FileInfo, string, error) {
	dirList := make(map[string]os.FileInfo)
	slnPath := ""

	err := filepath.Walk(dirPath, func(filePath string, fi os.FileInfo, walkErr error) error {
		// if an error occurred while walking the directory bubble up
		if walkErr != nil {
			return walkErr
		}

		// get the file path relative to the dir path
		rel := strings.Replace(filePath, dirPath, "", 1)

		// TODO: improve this section, this is a sloppy hack to ignore some binary directories
		// this could cause all kinds of issues if directory names are partial matches i.e.,
		// since 'packages' is ignored a directory called 'asset_packages' would be ignored.
		if len(rel) == 0 {
			// if we are at the root, or are looking at an ignored file skip it
			return nil
		} else if !fi.IsDir() && (ignoreFiles.MatchString(rel) || ignoreDirectories.MatchString(rel)) {
			// if this is a file, and it matches a file pattern, or contains an ignored directory skip it
			return nil
		} else if fi.IsDir() && ignoreDirectories.MatchString(rel) {
			// if this contains an ignored directory skip it
			return nil
		}

		ext := strings.ToLower(filepath.Ext(fi.Name()))

		if ext == ".sln" {
			slnPath = filePath
		}

		dirList[rel] = fi

		return nil
	})

	return dirList, slnPath, err
}

func (c *DotnetCommand) extractGUIDsFromSln(filePath string) (guids []string, err error) {
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return
	}

	guids, err = c.ExtractGUIDsFromString(string(contents))

	return
}

// ExtractGUIDsFromString extracts guids from file contents
func (c *DotnetCommand) ExtractGUIDsFromString(contents string) (guids []string, err error) {
	lines := strings.Split(contents, "\n")

	guids = []string{}

	for _, line := range lines {
		matches := guidInSln.FindStringSubmatch(strings.TrimSpace(line))

		if len(matches) > 1 {
			guids = append(guids, matches[1])
		}
	}

	return
}

// GeneratenewGUIDs takes a list of GUIDs and generates a new GUID to replaced
// each GUID in the list
func (c *DotnetCommand) GeneratenewGUIDs(guids []string) (guidMap map[string]string) {
	guidMap = make(map[string]string)

	for _, guid := range guids {
		guidMap[guid] = strings.ToUpper(uuid.NewV4().String())
	}

	return
}

func copyItems(dirList map[string]os.FileInfo, guids map[string]string) (err error) {
	fmt.Println("Generating new items...")

	// first we need to get our list of paths in order
	paths := []string{}

	for relPath := range dirList {
		paths = append(paths, relPath)
	}

	sort.Strings(paths)

	// now that we have the paths in order we can iteration and copy/transform the items
	for _, relPath := range paths {
		// transform the path to use the new solution name
		dstRel := strings.Replace(relPath, tplName, dstName, -1)

		tpl := path.Join(tplPath, relPath)
		dst := path.Join(dstPath, dstRel)

		err = copyItem(tpl, dst, dirList[relPath], guids)
		if err != nil {
			return err
		}

		fmt.Printf("  %s -> %s\n", relPath, dstRel)
	}

	return err
}

func copyItem(src, dst string, fi os.FileInfo, guids map[string]string) (err error) {
	if fi.IsDir() {
		if _, err := os.Stat(dst); os.IsNotExist(err) {
			os.Mkdir(dst, 0755)
		}
	} else {
		ext := strings.ToLower(filepath.Ext(fi.Name()))

		// if this file is a type known to contain project names,
		// edit it, otherwise just copy it
		if _, ok := dotnet.VsData.SourceFileExts[ext]; ok {
			err = transformFile(src, dst, guids)
		} else if _, ok := dotnet.VsData.ProjectFileExts[ext]; ok {
			err = transformFile(src, dst, guids)
		} else {
			err = copyFile(src, dst)
		}
	}

	return
}

func transformFile(src, dst string, guids map[string]string) (err error) {
	read, err := ioutil.ReadFile(src)
	if err != nil {
		return
	}

	// rename
	newContents := strings.Replace(string(read), tplName, dstName, -1)

	// replace guids
	if guids != nil {
		for k, v := range guids {
			newContents = strings.Replace(newContents, k, v, -1)
		}
	}

	err = ioutil.WriteFile(dst, []byte(newContents), 0666)

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
usage: ouroboros dotnet <args> [tpl path] [dst path]
  
Available args are:  
    new-guids     Generate new project GUIDs
	tpl-name      Override template solution name
`
	return strings.TrimSpace(helpText)
}
