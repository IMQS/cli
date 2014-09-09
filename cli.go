/*
Package cli is used to build the front-end of command-line applications

Commands have zero or more mandatory arguments, and zero or more optional parameters.
An example of a command invocation is:

	myapp -root=somedir -z -x thecommand param1 param2

Optional parameters can be boolean values, which are specified with a single hyphen:

	-z             This is a boolean option

Options parameters that can take on a value must be specified like this:

	-config=file   This is a value option

If no command is given, then the application help is displayed, for example:

	>imqsauth

	imqsauth -c=configfile [options] command

	  createdb         Create the postgres database
	  resetauthgroups  Reset the [admin,enabled] groups
	  createuser       Create a user in the authentication system

	  -c=configfile    Specify the authaus config file. A pseudo file called !TESTCONFIG1
	                   is used by the REST test suite to load a test configuration.
	                   This option is mandatory.

If one invokes help on a specific command, then details for that command are shown:

	>imqsauth help createuser

	createuser identity password

	  Create a user in the authentication system.
	  This affects only the 'authentication' system - the permit
	  database is not altered by this command.

	  -update   If specified, and the user already exists, then behave identically
	            to 'setpassword'. If this is not specified, and the identity
	            already exists, then the function returns with an error.

*/
package cli

import (
	"fmt"
	"os"
	"strings"
)

// An option to a command
type Option struct {
	Key         string // Must be present. Option is entered as -Key, or -Key=Value
	Value       string // If empty, then this is a boolean option, specified as -Key. If not empty, then this is a key/value option, specified as -Key=Value
	Description string
}

func findOption(options []Option, name string) *Option {
	for i := range options {
		if options[i].Key == name {
			return &options[i]
		}
	}
	return nil
}

// Command execution callback.
// It is often easiest to implement a number of commands as a single big function with a switch statement on 'cmd'.
type ExecFunc func(cmd string, args []string, options map[string]string)

// A top-level command
type Command struct {
	Name        string
	Description string
	Args        []string // Mandatory arguments. To specify a variable number of arguments, write "...values" on the last argument. The name after the three dots can be anything.
	Options     []Option // Optional arguments
	Exec        ExecFunc // If this is nil, then App.DefaultExec is called
}

func isVarArgs(args []string) bool {
	return len(args) != 0 && strings.Index(args[len(args)-1], "...") == 0
}

// The contents of Description, before the first \n
func (c *Command) ShortDescription() string {
	return strings.Split(c.Description, "\n")[0]
}

// The contents of Description, after the first \n
func (c *Command) ExtraDescription() string {
	return strings.Join(strings.Split(c.Description, "\n")[1:], "\n")
}

// Add a command-specific bool option (such as -z)
func (c *Command) AddBoolOption(name, description string) {
	opt := Option{
		Key:         name,
		Description: description,
	}
	c.Options = append(c.Options, opt)
}

// Add a command-specific value option (such as -c=config_file)
func (c *Command) AddValueOption(name, value, description string) {
	opt := Option{
		Key:         name,
		Value:       value,
		Description: description,
	}
	c.Options = append(c.Options, opt)
}

// Application
type App struct {
	Description string     // Single-line description
	DefaultExec ExecFunc   // Exec callback that is used if command's Exec is nil
	Commands    []*Command // Commands
	Options     []Option   // Global options
}

// Add an application-wide bool option (such as -z)
func (app *App) AddBoolOption(name, description string) {
	opt := Option{
		Key:         name,
		Description: description,
	}
	app.Options = append(app.Options, opt)
}

// Add an application-wide value option (such as -c=config_file)
func (app *App) AddValueOption(name, value, description string) {
	opt := Option{
		Key:         name,
		Value:       value,
		Description: description,
	}
	app.Options = append(app.Options, opt)
}

// Execute a command list.
func (app *App) Run() {
	options := map[string]string{}
	cmdName := ""
	cmdArgs := []string{}
	for iarg, arg := range os.Args {
		if iarg == 0 {
			// executable name
			continue
		}
		if arg[0:1] == "-" {
			equals := strings.Index(arg, "=")
			if equals != -1 {
				options[arg[1:equals]] = arg[equals+1:]
			} else {
				options[arg[1:]] = ""
			}
		} else {
			if cmdName == "" {
				cmdName = arg
			} else {
				cmdArgs = append(cmdArgs, arg)
			}
		}
	}

	_, haveHelpOption := options["help"]
	if cmdName == "" || cmdName == "help" || haveHelpOption {
		//fmt.Printf("cmdArgs = %v\n", strings.Join(cmdArgs, ","))
		if len(cmdArgs) >= 1 {
			app.ShowHelp(cmdArgs[0])
		} else {
			app.ShowHelp(cmdName)
		}
		return
	}

	cmd := app.find(cmdName)
	if cmd != nil {
		allOptions := append(app.Options, cmd.Options...)
		isVArgs := isVarArgs(cmd.Args)
		if isVArgs {
			if len(cmdArgs) < len(cmd.Args)-1 {
				fmt.Printf("%v arguments given, but %v needs '%v'\n", len(cmdArgs), cmdName, formatCmdArgs(cmd.Args))
				return
			}
		} else if len(cmdArgs) != len(cmd.Args) {
			fmt.Printf("%v arguments given, but %v needs '%v'\n", len(cmdArgs), cmdName, formatCmdArgs(cmd.Args))
			return
		}
		for key, value := range options {
			opt := findOption(allOptions, key)
			if opt == nil {
				fmt.Printf("Unrecognized option %v\n", key)
				return
			} else if (opt.Value == "") && (value != "") {
				fmt.Printf("Option %v does not take a value. Simply use -%v\n", opt.Key, opt.Key)
				return
			} else if (opt.Value != "") && (value == "") {
				fmt.Printf("Option %v needs a value. Use -%v=%v\n", opt.Key, opt.Key, opt.Value)
				return
			}
		}
		exec := cmd.Exec
		if exec == nil {
			exec = app.DefaultExec
		}
		if exec == nil {
			fmt.Printf("No exec function specified for command '%v'\n", cmdName)
			return
		}
		exec(cmdName, cmdArgs, options)
	} else {
		fmt.Printf("Unrecognized command '%v'\n", cmdName)
	}
}

func (app *App) AddCommand(name, description string, args ...string) *Command {
	cmd := &Command{
		Name:        name,
		Description: description,
		Args:        args,
	}
	app.Commands = append(app.Commands, cmd)
	return cmd
}

func (app *App) find(cmdName string) *Command {
	for i := range app.Commands {
		if app.Commands[i].Name == cmdName {
			return app.Commands[i]
		}
	}
	return nil
}

func formatTextIntoLines(text string, firstLineIndent, otherLinesIndent int) []string {
	const width = 55
	words := strings.Split(text, " ")
	line := ""
	lines := []string{}
	for i := range words {
		line += words[i]
		if len(line) > width {
			lines = append(lines, line)
			line = ""
		} else if i != len(words)-1 {
			line += " "
		}
	}
	if len(line) != 0 {
		lines = append(lines, line)
	}
	return lines
}

func writeBody(text string, firstLineIndent, otherLinesIndent int) {
	lines := formatTextIntoLines(text, firstLineIndent, otherLinesIndent)
	for i, line := range lines {
		if i == 0 {
			fmt.Printf("%v%v\n", strings.Repeat(" ", firstLineIndent), line)
		} else {
			fmt.Printf("%v%v\n", strings.Repeat(" ", otherLinesIndent), line)
		}
	}
}

func formatCmdArgs(args []string) string {
	if isVarArgs(args) {
		varg := args[len(args)-1][3:]
		return strings.Join(args[0:len(args)-1], " ") + " " + varg + "1 " + varg + "2 " + varg + "3..."
	} else {
		return strings.Join(args, " ")
	}
}

// This is called automatically by Run().
func (app *App) ShowHelp(cmdName string) {

	findLongestOption := func(options []Option) int {
		max := 0
		for _, opt := range options {
			length := 0
			if opt.Value != "" {
				length = len(opt.Key) + 1 + len(opt.Value)
			} else {
				length = len(opt.Key)
			}
			if length > max {
				max = length
			}
		}
		return max
	}

	optionFormatStr := ""

	formatOption := func(opt Option) string {
		if opt.Value != "" {
			pair := fmt.Sprintf("%v=%v", opt.Key, opt.Value)
			return fmt.Sprintf(optionFormatStr, pair)
		} else {
			return fmt.Sprintf(optionFormatStr, opt.Key)
		}
	}

	showOptions := func(options []Option) {
		longest := findLongestOption(options)
		optionFormatStr = fmt.Sprintf("  -%%-%vv", longest)
		fmt.Printf("\n")
		for _, opt := range options {
			fmt.Printf("%v", formatOption(opt))
			writeBody(opt.Description, 3, 6+longest)
		}
	}

	cmd := app.find(cmdName)
	if cmd != nil {
		cmdAndArgs := cmd.Name + " " + formatCmdArgs(cmd.Args)
		fmt.Printf("\n%v\n\n", cmdAndArgs)
		writeBody(cmd.ShortDescription()+".", 2, 2)
		if cmd.ExtraDescription() != "" {
			writeBody(cmd.ExtraDescription(), 2, 2)
		}
		if len(cmd.Options) != 0 {
			showOptions(cmd.Options)
		}
	} else {
		longestCmd := 0
		for _, c := range app.Commands {
			if len(c.Name) > longestCmd {
				longestCmd = len(c.Name)
			}
		}
		cmdFormatStr := fmt.Sprintf("  %%-%vv  %%v\n", longestCmd)
		if app.Description != "" {
			fmt.Printf("\n%v\n\n", app.Description)
		}
		for _, c := range app.Commands {
			fmt.Printf(cmdFormatStr, c.Name, c.ShortDescription())
		}
		if len(app.Options) != 0 {
			showOptions(app.Options)
		}
	}

}
