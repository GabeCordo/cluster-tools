package client

import (
	"fmt"
	"github.com/GabeCordo/commandline"
	"github.com/GabeCordo/etl/core"
	"github.com/GabeCordo/etl/utils/template"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// CREATE PROJECT COMMAND START

type CreateProjectCommand struct {
	PubicName string
}

func (cproj CreateProjectCommand) Name() string {
	return cproj.PubicName
}

func (cproj CreateProjectCommand) Run(cl *commandline.CommandLine) commandline.TerminateOnCompletion {
	if cl.Flags.Create {
		cproj.CreateProject(cl)
	} else if cl.Flags.Delete {
		cproj.DeleteProject(cl)
	} else if cl.Flags.Show {
		cproj.ShowProjects(cl)
	}

	return true // complete
}

func (cproj CreateProjectCommand) CreateProject(cl *commandline.CommandLine) commandline.TerminateOnCompletion {

	if !cl.Flags.Create {
		log.Println("(!) missing create flag")
		return true
	}

	var projectName string
	if projectName = cl.NextArg(); projectName == commandline.FinalArg {
		log.Println("(!) missing project Name")
		return true
	}

	projectPath := commandline.EmptyPath().Dir(cl.MetaData.WorkingDirectory).Dir(projectName)

	if err := projectPath.MkDir(); err != nil {
		log.Println("(!) failed to create project directory, a project of that PublicName likely already exists in this directory")
		return true
	}

	projectConfig := core.NewConfig(projectName)

	projectConfigPath := projectPath.File("config.etl.json")
	projectConfig.ToJson(projectConfigPath.ToString())

	// GENERATE THE DEFAULT PROJECT FILES

	stringRepOfModFile := fmt.Sprintf(
		"module %s\n\ngo %s\n",
		projectName,
		runtime.Version()[2:])
	dependencies := []string{"github.com/GabeCordo/commandline v0.1.1", "github.com/GabeCordo/fack v0.1.2"}
	if len(dependencies) > 0 {
		stringRepOfModFile += "\nrequire("

		for _, dependency := range dependencies {
			stringRepOfModFile += fmt.Sprintf("\n\t%s", dependency)
		}

		stringRepOfModFile += "\n)\n"
	}

	goModFilePath := projectPath.File("go.mod")
	projectGoModulePath, err := os.Create(goModFilePath.ToString())
	if err != nil {
		log.Print("(!) failed to create go module")
		return true
	}

	if _, err := projectGoModulePath.Write([]byte(stringRepOfModFile)); err != nil {
		log.Println("(!) failed to " +
			"create go module")
		return true
	}

	var processedRootTemplateFile []byte
	projectRootTemplatePath := TemplateFolder().File("project.root.go")
	if bytes, err := projectRootTemplatePath.Read(); err == nil {
		match := make(map[string]string)
		match["project"] = projectName
		processedRootTemplateFile = template.Process(bytes, match)
	} else {
		log.Println("(!) a template file is missing")
		return true
	}

	rootGoFilePath := projectPath.File(projectName + ".root.go")
	os.Create(rootGoFilePath.ToString())
	if err := rootGoFilePath.Write(processedRootTemplateFile); err != nil {
		log.Println("(!) failed to create root project go file")
		return true
	}

	projectSrcDirPath := projectPath.Dir("src")
	if err := projectSrcDirPath.MkDir(); err != nil {
		log.Println("(!) failed to create project files")
		return true
	}

	vectorEtlTemplatePath := TemplateFolder().File("vector.etl.go")
	if exampleVectorEtlBytes, err := vectorEtlTemplatePath.Read(); err == nil {

		vectorEtlProjectPath := projectPath.Dir("src").File("vector.etl.go")
		os.Create(vectorEtlProjectPath.ToString())
		if err = vectorEtlProjectPath.Write(exampleVectorEtlBytes); err != nil {
			log.Println("(!) failed writing default Vector example to project")
			return true
		}
	} else {
		log.Println("(!) failed reading default Vector example from .templates")
		return true
	}

	projectTestDirectoryPath := projectPath.Dir("test")
	if err := projectTestDirectoryPath.MkDir(); err != nil {
		fmt.Println("(!) failed to create project files")
		return true
	}

	// INSTALL THE GO MODULES REQUIRED BY THE TEMPLATE

	cmd := exec.Command("go mod tidy")
	if err := cmd.Run(); err != nil {
		log.Println("(!) failed to go mod tidy, you will need to install the dependencies manually")
	}

	// ADD THE PROJECT TO THE GLOBAL CLI CONFIG

	projectsConfigFilePath := commandline.EmptyPath().Dir(RootEtlFolder()).File("projects.etl.json")
	if projectsConfigFilePath.DoesNotExist() {
		fmt.Println("the etl install is corrupted, you are missing a projects.elt.json file")
		return true
	}

	clientConfig := JSONToConfig(projectsConfigFilePath)
	clientConfig.AddProject(Project{
		Name:      projectName,
		Directory: projectPath.ToString(),
		CreatedOn: time.Now(),
	})

	err = clientConfig.ToJson(projectsConfigFilePath)
	if err != nil {
		log.Println("(warning) failed to add the project to the global records")
	}

	return true
}

func (cproj CreateProjectCommand) DeleteProject(cl *commandline.CommandLine) commandline.TerminateOnCompletion {

	projectsConfigFilePath := commandline.EmptyPath().Dir(RootEtlFolder()).File("projects.etl.json")
	if projectsConfigFilePath.DoesNotExist() {
		fmt.Println("the etl install is corrupted, you are missing a projects.elt.json file")
		return true
	}

	projectName := cl.NextArg()
	if projectName == commandline.FinalArg {
		fmt.Println("missing project PublicName to delete")
		return true
	}

	clientConfig := JSONToConfig(projectsConfigFilePath)

	projectFound := false
	modifiedProjectsList := make([]Project, 0)
	for _, project := range clientConfig.Projects {
		if project.Name == projectName {
			projectFound = true
		} else {
			modifiedProjectsList = append(modifiedProjectsList, project)
		}
	}

	if !projectFound {
		fmt.Printf("no project with the PublicName %s exists\n", projectName)
		return true
	}

	clientConfig.Projects = modifiedProjectsList // a slice without the deleted project

	err := clientConfig.ToJson(projectsConfigFilePath)
	if err != nil {
		log.Println("(warning) failed to add the project to the global records")
	}

	return true
}

func (cproj CreateProjectCommand) ShowProjects(cl *commandline.CommandLine) commandline.TerminateOnCompletion {

	projectsConfigFilePath := commandline.EmptyPath().Dir(RootEtlFolder()).File("projects.etl.json")
	if projectsConfigFilePath.DoesNotExist() {
		fmt.Println("the etl install is corrupted, you are missing a projects.elt.json file")
		return true
	}

	clientConfig := JSONToConfig(projectsConfigFilePath)

	if len(clientConfig.Projects) == 0 {
		fmt.Println("no etl projects on the local system")
		return true
	}

	fmt.Println("etl projects on the local system:")
	for _, project := range clientConfig.Projects {
		fmt.Printf("\nName: %s\nCreated On: %s\nDirectory %s\n", project.Name, project.CreatedOn, project.Directory)
	}

	return true
}
