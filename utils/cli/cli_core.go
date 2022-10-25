package cli

import (
	"etl/core"
	"fmt"
	"os"
	"strings"
	"time"
)

func NewCommandLine(core *core.Core) (*CommandLine, bool) {
	if core == nil {
		return nil, false
	}

	cli := new(CommandLine)
	cli.Core = core

	// make Command map
	cli.Commands = make(map[string]Command)

	// help commands
	cli.AddFunction([]string{"version"}, VersionCommand{"version"})
	cli.AddFunction([]string{"h", "help"}, HelpCommand{"help"})

	// cli flag sub-commands
	cli.AddFunction([]string{"c", "create"}, CreateCommand{"create"})
	cli.AddFunction([]string{"d", "delete"}, DeleteCommand{"delete"})
	cli.AddFunction([]string{"s", "show"}, ShowCommand{"show"})

	// cli core commands
	cli.AddFunction([]string{"key"}, GenerateKeyPairCommand{"keys"})
	cli.AddFunction([]string{"project"}, CreateProjectCommand{"create-project"})
	cli.AddFunction([]string{"cluster"}, CreateClusterCommand{"create-cluster"})
	cli.AddFunction([]string{"profile"}, ProfileCommand{"create-profile"})

	// not complete
	cli.AddFunction([]string{"i", "interactive"}, InteractiveDashboardCommand{"dashboard"})

	// to run the etl backend
	cli.AddFunction([]string{"debug"}, DebugCommand{"debug"})
	cli.AddFunction([]string{"deploy"}, &DeployCommand{"deploy"})

	// retrieve the config that should be located in the root directory of the project
	cli.Config = new(Config)
	JSONToCLIConfig(cli.Config)

	return cli, true
}

func (cli *CommandLine) AddFunction(identifiers []string, function Command) {
	// loop over every identifier the developer wants to associate with the function
	// and ensure that no mapping already exists for the CommandLine instance
	//
	// Why wouldn't we just reject the collided identifiers and bind the one's
	// that remain?
	//
	// This would provide the developer with a false sense of success and could
	// open up the code to misinterpretation. The developer is not getting any
	// feedback on what identifiers are being accepted and even if we did provide
	// that back, what says they are verifying this information?
	//
	for _, identifier := range identifiers {
		// if there is a collision, tell the developer exactly why the collision occurred
		// in-case they accidentally gave another function an unintended identifier
		if otherFunction, found := cli.Commands[identifier]; found {
			st := "cannot bind %s to %s as an association already exists for %s"
			s := fmt.Sprintf(st, identifier, function.Name(), otherFunction.Name())
			panic(s)
		}
	}

	for _, identifier := range identifiers {
		cli.Commands[identifier] = function
	}
}

func (cli *CommandLine) NextArg() string {
	cli.argsPointer++

	if cli.argsPointer < cli.numOfArgs {
		return cli.args[cli.argsPointer]
	} else {
		return FinalArg
	}
}

func (cli *CommandLine) CollectMetaData() {
	// exe time start
	cli.MetaData.TimeCalled = time.Now()
	// exe time end

	// system path start
	//executable, err := os.Executable()
	//if err != nil {
	//	panic("error 00")
	//}
	cli.MetaData.WorkingDirectory, _ = os.Getwd()
	// system path end
}

// UnknownCommand
// @runtime O(keywords * chars)
func (cli CommandLine) UnknownCommand(identifier string) Terminate {

	// populate a map with all keywords where the first char of the keyword matches the first char of the identifier
	// O(keywords)
	keywords := make(map[string]int)
	for keyword, _ := range cli.Commands {
		if keyword[0] == identifier[0] {
			keywords[keyword] = 1
		}
	}

	// whatever word has the greatest number of chars common with the unknown identifier will be outputted
	// if there is more than one identifier that shared the number of common chars, output them all
	mostNumOfCharOccurrences := 1

	// start at the second index since we know the first one is already common
	for c := 1; c < len(identifier); c++ {
		char := string(identifier[c])
		for keyword, _ := range keywords {
			if strings.Contains(keyword, char) {
				keywords[keyword]++

				if keywords[keyword] > mostNumOfCharOccurrences {
					mostNumOfCharOccurrences = keywords[keyword]
				}
			}
		}
	}

	var output string
	if len(keywords) == 0 {
		output = "Unknown identifier \"%s\", there are no matching commands.\n"
	} else {
		output = "Unknown identifier \"%s\", did you mean?\n"
	}
	fmt.Printf(output, identifier)

	// if keywords exist in the map, for loop entered, otherwise it will skip (this is why
	// we do not need to worry about averageOfCharOccurrences being initialized if numOfKeywords is zero
	for keyword, count := range keywords {
		//if count >= averageOfCharOccurrences {
		if count == mostNumOfCharOccurrences {
			fmt.Println(keyword)
		}
	}

	return true // terminate since we don't know how to execute all commands
}

func (cli *CommandLine) Run() {

	// this data can be useful for capturing the environment the exe was called
	cli.CollectMetaData()

	// push the command line arguments into struct
	cli.args = os.Args[1:] // strip out the file descriptor in position 0
	cli.numOfArgs = len(cli.args)
	cli.argsPointer = 0

	for arg := cli.args[0]; arg != FinalArg; arg = cli.NextArg() {
		var terminateFlag Terminate = false
		if command, found := cli.Commands[arg]; found {
			terminateFlag = command.Run(cli)
		} else {
			terminateFlag = cli.UnknownCommand(arg)
		}

		if terminateFlag {
			return
		}
	}
	// stop reading cli arguments

}
