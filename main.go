package main

import (
	"fmt"
	"github.com/GabeCordo/commandline"
	"os"
	"time"
)

func Version(commandLine *commandline.CommandLine) string {
	strVersion := fmt.Sprintf("%.2f", commandLine.Config.Version)
	strTimeNow := time.Now().Format("Mon Jan _2 15:04:05 MST 2006")
	return "ETLFramework Version " + strVersion + " " + strTimeNow
}

func InitializeFolders() (commandline.Path, commandline.Path) {

	executablePathStr, _ := os.Executable()
	executablePath := commandline.EmptyPath().Dir(executablePathStr)

	dataFolderPath := executablePath.Dir(".data")

	if dataFolderPath.DoesNotExist() {
		dataFolderPath.MkDir()
	}

	templateFolderPath := executablePath.BackDir().Dir("..templates")

	if templateFolderPath.DoesNotExist() {
		panic("template path missing")
	}

	return dataFolderPath, templateFolderPath
}

func TemplateFolderPath() commandline.Path {
	executablePath, _ := os.Executable()
	return commandline.EmptyPath().Dir(executablePath).Dir(".bin").Dir(".templates")
}

func main() {

	//dataFolderPath, templateFolderPath := InitializeFolders()

	if commandLine := commandline.NewCommandLine(); commandLine != nil {

		commandLine.AddCommand([]string{"version"}, VersionCommand{"version"})

		// cli core commands
		commandLine.AddCommand([]string{"key"}, GenerateKeyPairCommand{"keys"})
		commandLine.AddCommand([]string{"project"}, CreateProjectCommand{"create-project"})
		commandLine.AddCommand([]string{"cluster"}, CreateClusterCommand{"create-cluster"})
		commandLine.AddCommand([]string{"profile"}, ProfileCommand{"create-profile"})

		//local project interaction
		commandLine.AddCommand([]string{"deploy"}, DeployCommand{"deploy"})

		// remote project interaction

		// remote project observation (not complete)
		commandLine.AddCommand([]string{"i", "interactive"}, InteractiveDashboardCommand{"dashboard"})

		commandLine.Run()
	}
}
