// This example demonstrates a command-line application
package cli_test

import (
	"fmt"
	"github.com/IMQS/cli"
)

func exec(name string, args []string, options map[string]string) {
	switch name {
	case "start":
		fmt.Printf("starting in %v on port %v.\n", args[0], args[1])
	case "clean":
		fmt.Printf("cleaning %v with %v strength\n", args[0], options["strength"])
	}
}

func Example_application() {
	app := cli.App{}
	app.Description = "myapp [options] command"
	app.DefaultExec = exec

	app.AddCommand("start", "Start the application", "port", "root-directory")

	// Add a command with some optional arguments
	init := app.AddCommand("initialize", "Initialize a directory\nThis will setup the necessary structures in 'directory'.", "root-directory")
	init.AddBoolOption("clean", "Clean all files")
	init.AddValueOption("strength", "howstrong", "Strong values clean more files")

	// This command takes one mandatory argument, followed by zero or more variable arguments
	app.AddCommand("varargs", "Demonstrate variable number of arguments", "param1", "...things")

	app.Run()
}
