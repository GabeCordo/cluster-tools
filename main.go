package main

import (
	"etl/core"
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

func InitializeFolders() (utils.Path, utils.Path) {

	executablePathStr, _ := os.Executable()
	executablePath := utils.EmptyPath().Dir(executablePathStr)

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

func TemplateFolderPath() Path {
	executablePath, _ := os.Executable()
	return EmptyPath().Dir(executablePath).Dir(".bin").Dir(".templates")
}

func main() {

	dataFolderPath, templateFolderPath := InitializeFolders()

	c := core.NewCore()
	if commandLine, ok := cli.NewCommandLine(c); ok {

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
