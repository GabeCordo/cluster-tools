package client

import (
	"bufio"
	"fmt"
	"github.com/GabeCordo/commandline"
	"os"
)

// DEVELOPER PROFILE COMMAND START

type ProfileCommand struct {
	PublicName string
}

func (pc ProfileCommand) Name() string {
	return pc.PublicName
}

func (pc ProfileCommand) Run(cl *commandline.CommandLine) commandline.TerminateOnCompletion {

	if cl.Config == nil {
		fmt.Println("(!) The ETL Config is Corrupted")
		return true
	}

	if cl.Flags.Create {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("First Name: ")
		line, err := reader.ReadString('\n')
		if err != nil {
			panic("error reading profile")
		}
		cl.Config.UserProfile.FirstName = line[:len(line)-1] // remove the delim

		fmt.Print("Last Name: ")
		line, err = reader.ReadString('\n')
		if err != nil {
			panic("error reading profile")
		}
		cl.Config.UserProfile.LastName = line[:len(line)-1] // remove the delim

		fmt.Print("Email: ")
		line, err = reader.ReadString('\n')
		if err != nil {
			panic("error reading profile")
		}
		cl.Config.UserProfile.Email = line[:len(line)-1] // remove the delim

		cliConfigPath := commandline.EmptyPath().File("config.cli.json")
		cl.Config.ToJson(cliConfigPath) // push the JSON update to the local file
	} else if cl.Flags.Show {
		if (len(cl.Config.UserProfile.FirstName) == 0) && (len(cl.Config.UserProfile.LastName) == 0) && (len(cl.Config.UserProfile.Email) == 0) {
			fmt.Println("developer profile not configured")
			fmt.Println("use \"etl create profile\" to create a new developer profile")
		} else {
			fmt.Println(Version(cl))
			fmt.Println(cl.Config.UserProfile.FirstName + " " + cl.Config.UserProfile.LastName)
			fmt.Println(cl.Config.UserProfile.Email)
		}
	} else {
		fmt.Println("missing flag directive")
	}

	return true
}
