// This example demonstrates a command-line application
package cli_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/IMQS/cli"
)

func exec(name string, args []string, options cli.OptionSet) int {
	switch name {
	case "start":
		fmt.Printf("starting in %v on port %v.\n", args[0], args[1])
	case "initialize":
		fmt.Printf("initializing %v with %v strength. clean=%v\n", args[0], options["strength"], options.Has("clean"))
	case "varargs":
		fmt.Printf("defining varargs %v for %v", args, name)
	default:
		return 1
	}
	return 0
}

func TestExampleApplication(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	app := cli.App{}
	app.Description = "myapp [options] command"
	app.DefaultExec = exec

	app.AddCommand("start", "Start the application", "port", "root-directory")

	// Add a command with some optional arguments
	initialize := app.AddCommand("initialize", "Initialize a directory\nThis will setup the necessary structures in 'directory'.", "root-directory")
	initialize.AddBoolOption("clean", "Clean all files")
	initialize.AddValueOption("strength", "howstrong", "Strong values clean more files")

	// This command takes one mandatory argument, followed by zero or more variable arguments
	app.AddCommand("varargs", "Demonstrate variable number of arguments", "param1", "...things")

	os.Args = []string{"myapp", "start", "/folder1/folder2", "8669"}
	if app.Run() == 1 {
		t.Fatal(`command "start" has failed`)
	}

	os.Args = []string{"myapp", "initialize", "-clean", "strength=2"}
	if app.Run() == 1 {
		t.Fatal(`command "initialize" has failed`)
	}

	os.Args = []string{"myapp", "varargs", "param1", "param2", "param3"}
	if app.Run() == 1 {
		t.Fatal(`command "varargs" has failed`)
	}
}
