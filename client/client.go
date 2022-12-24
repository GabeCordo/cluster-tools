package client

import (
	"github.com/GabeCordo/commandline"
	"github.com/GabeCordo/etl/client/controllers"
	"github.com/GabeCordo/etl/client/core"
)

func CommandLineClient() {
	// these are data files used by this executable to store metadata about created projects or ECDSA keys
	core.IfMissingInitializeFolders()

	profilePath := core.CliConfigFile()
	if commandLine := commandline.NewCommandLine(profilePath); commandLine != nil {

		commandLine.AddCommand([]string{"version"}, controllers.VersionCommand{PublicName: "version"})

		// cli core commands
		commandLine.AddCommand([]string{"key"}, controllers.KeyPairCommand{PublicName: "keys"})
		commandLine.AddCommand([]string{"project"}, controllers.CreateProjectCommand{"create-project"})
		commandLine.AddCommand([]string{"cluster"}, controllers.ClusterCommand{PublicName: "cluster"})
		commandLine.AddCommand([]string{"profile"}, controllers.ProfileCommand{PublicName: "create-profile"})

		//local project interaction
		commandLine.AddCommand([]string{"deploy"}, controllers.DeployCommand{PublicName: "deploy"})
		commandLine.AddCommand([]string{"mount"}, controllers.MountCommand{PublicName: "mount"})
		commandLine.AddCommand([]string{"endpoint"}, controllers.EndpointCommand{PublicName: "endpoint"})
		commandLine.AddCommand([]string{"permission"}, controllers.PermissionCommand{PublicName: "permission"})

		// remote project interaction

		// remote project observation (not complete)
		commandLine.AddCommand([]string{"i", "interactive"}, controllers.InteractiveDashboardCommand{"dashboard"})

		commandLine.Run()
	}
}
