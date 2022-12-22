package main

import (
	"github.com/GabeCordo/commandline"
	"github.com/GabeCordo/etl/client"
)

func main() {

	//dataFolderPath, templateFolderPath := InitializeFolders()

	profilePath := commandline.EmptyPath().Dir(client.RootEtlFolder()).File("config.cli.json")
	if commandLine := commandline.NewCommandLine(profilePath); commandLine != nil {

		commandLine.AddCommand([]string{"version"}, client.VersionCommand{"version"})

		// cli core commands
		commandLine.AddCommand([]string{"key"}, client.GenerateKeyPairCommand{"keys"})
		commandLine.AddCommand([]string{"project"}, client.CreateProjectCommand{"create-project"})
		commandLine.AddCommand([]string{"cluster"}, client.ClusterCommand{"cluster"})
		commandLine.AddCommand([]string{"profile"}, client.ProfileCommand{"create-profile"})

		//local project interaction
		commandLine.AddCommand([]string{"deploy"}, client.DeployCommand{"deploy"})
		commandLine.AddCommand([]string{"mount"}, client.MountCommand{"mount"})

		// remote project interaction

		// remote project observation (not complete)
		commandLine.AddCommand([]string{"i", "interactive"}, client.InteractiveDashboardCommand{"dashboard"})

		commandLine.Run()
	}
}
