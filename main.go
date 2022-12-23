package main

import (
	"github.com/GabeCordo/commandline"
	"github.com/GabeCordo/etl/client"
)

func main() {

	// these are data files used by this executable to store metadata about created projects or ECDSA keys
	client.IfMissingInitializeFolders()

	profilePath := client.CliConfigFile()
	if commandLine := commandline.NewCommandLine(profilePath); commandLine != nil {

		commandLine.AddCommand([]string{"version"}, client.VersionCommand{PublicName: "version"})

		// cli core commands
		commandLine.AddCommand([]string{"key"}, client.KeyPairCommand{PublicName: "keys"})
		commandLine.AddCommand([]string{"project"}, client.CreateProjectCommand{"create-project"})
		commandLine.AddCommand([]string{"cluster"}, client.ClusterCommand{PublicName: "cluster"})
		commandLine.AddCommand([]string{"profile"}, client.ProfileCommand{PublicName: "create-profile"})

		//local project interaction
		commandLine.AddCommand([]string{"deploy"}, client.DeployCommand{PublicName: "deploy"})
		commandLine.AddCommand([]string{"mount"}, client.MountCommand{PublicName: "mount"})

		// remote project interaction

		// remote project observation (not complete)
		commandLine.AddCommand([]string{"i", "interactive"}, client.InteractiveDashboardCommand{"dashboard"})

		commandLine.Run()
	}
}
