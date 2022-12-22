package client

import (
	"fmt"
	"github.com/GabeCordo/commandline"
	"os"
	"time"
)

func RootEtlFolder() string {
	executableFilePath, _ := os.Executable()
	return executableFilePath[:len(executableFilePath)-10] // remove "/build/etl" from the end of the path
}

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

// HELPER COMMANDS

func TemplateFolder() commandline.Path {
	rootEtlFolder := RootEtlFolder()
	return commandline.EmptyPath().Dir(rootEtlFolder).Dir(".templates")
}
