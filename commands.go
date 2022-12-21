package main

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"fmt"
	"github.com/GabeCordo/commandline"
	"github.com/GabeCordo/etl/core"
	"github.com/GabeCordo/etl/utils/template"
	"github.com/GabeCordo/fack"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unicode"
	"unicode/utf8"
)

// HELPER COMMANDS

func TemplateFolder() commandline.Path {
	rootEtlFolder := RootEtlFolder()
	return commandline.EmptyPath().Dir(rootEtlFolder).Dir(".templates")
}

// VERSION COMMAND START

type VersionCommand struct {
	name string
}

func (vc VersionCommand) Name() string {
	return vc.name
}

func (vc VersionCommand) Run(cl *commandline.CommandLine) commandline.TerminateOnCompletion {
	fmt.Println(Version(cl))
	return true
}

// GENERATE KEY PAIR START

type GenerateKeyPairCommand struct {
	name string
}

func (gkpc GenerateKeyPairCommand) Name() string {
	return gkpc.name
}

func (gkpc GenerateKeyPairCommand) Run(cl *commandline.CommandLine) commandline.TerminateOnCompletion {

	// we only want to create a key if it has been pre-pended with the 'create' flag
	if cl.Flags.Create {
		// generate a public / private key pair
		pair, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			fmt.Println("Could not generate public and private key pair")
			return true
		}

		x509Encoded, _ := x509.MarshalECPrivateKey(pair)
		fmt.Println("[private]")
		fmt.Println(fack.ByteToString(x509Encoded))

		x509EncodedPub, err := x509.MarshalPKIXPublicKey(&pair.PublicKey)
		fmt.Println(len(x509EncodedPub))
		fmt.Println("[public]")
		fmt.Println(fack.ByteToString(x509EncodedPub))
	} else {
		fmt.Println("key specified without an action [create/delete]?")
	}

	return true // this is a terminal command
}

// INTERACTIVE DASHBOARD START

type InteractiveDashboardCommand struct {
	name string
}

func (idc InteractiveDashboardCommand) Name() string {
	return idc.name
}

func (idc InteractiveDashboardCommand) Run(cl *commandline.CommandLine) commandline.TerminateOnCompletion {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs // block until we receive an interrupt from the system
		fmt.Println()
		os.Exit(0)
	}()

	for {
		now := time.Now()
		fmt.Printf("%d:%d:%d\r", now.Hour(), now.Minute(), now.Second())

		time.Sleep(1 * time.Second)
	}

	return true
}

// DEPLOY COMMAND START

type DeployCommand struct {
	name string
}

func (dc DeployCommand) Name() string {
	return dc.name
}

func (dc DeployCommand) Run(cl *commandline.CommandLine) commandline.TerminateOnCompletion {

	if cl.Flags.Debug {
		log.Println("(+) starting up etl")
	}

	mainPath := commandline.EmptyPath().File("main.go")
	if _, err := os.Stat(mainPath.ToString()); err == nil {
		// if the file exists run the main module
		runEtlMainCmd := exec.Command("go run main.go")
		if err = runEtlMainCmd.Run(); err != nil {
			// there was an source error inside the etl project
			log.Print(err)
		}
	} else {
		// if the file does not exists, let them know that they are not in an etl project folder
		log.Println("(!) you are not in an ETL project")
	}

	if cl.Flags.Debug {
		log.Println("(-) shutting down etl")
	}

	return true // end of program
}

// CREATE PROJECT COMMAND START

type CreateProjectCommand struct {
	name string
}

func (cproj CreateProjectCommand) Name() string {
	return cproj.name
}

func (cproj CreateProjectCommand) Run(cl *commandline.CommandLine) commandline.TerminateOnCompletion {

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
		log.Println("(!) failed to create project directory, a project of that name likely already exists in this directory")
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

	return true
}

// CLUSTER COMMAND START

type ClusterCommand struct {
	name string
}

func (cc ClusterCommand) Name() string {
	return cc.name
}

func (cc ClusterCommand) Run(cl *commandline.CommandLine) commandline.TerminateOnCompletion {

	if cl.Flags.Create {
		return cc.CreateCluster(cl)
	} else if cl.Flags.Delete {
		return cc.DeleteCluster(cl)
	} else if cl.Flags.Show {
		return cc.ShowClusters(cl)
	} else {
		return true
	}

}

func (cc ClusterCommand) CreateCluster(cl *commandline.CommandLine) commandline.TerminateOnCompletion {
	projectPath := commandline.EmptyPath().Dir(cl.MetaData.WorkingDirectory)

	// see if we are currently in an etl project, otherwise we cannot add a cluster
	configPath := projectPath.File("config.etl.json")
	if !configPath.Exists() {
		fmt.Println("no etl project exists")
		return true
	}

	// read the config file
	var projectConfig core.Config
	core.JSONToETLConfig(&projectConfig, configPath.ToString())

	// the cluster Name should be stored as the second argument
	clusterName := cl.NextArg()
	if clusterName == commandline.FinalArg {
		fmt.Println("missing cluster Name")
		return true
	}

	// all clusters must start with a capitol letter and have a length of at least three
	utf8ClusterName, _ := utf8.DecodeRuneInString(clusterName)
	if unicode.IsLower(utf8ClusterName) {
		fmt.Println("cluster must start with an uppercase letter")
		return true
	} else if len(clusterName) < 3 {
		fmt.Println("cluster must be at least 3 characters long")
		return true
	}

	// Collect Metadata about the creator

	var firstName, lastName, email string

	if cl.Config == nil {
		fmt.Println("(!) you are seeing this because your etl profile is missing")

		fmt.Print("First Name: ")
		fmt.Sprintln(&firstName)
		fmt.Println()

		fmt.Print("Last Name: ")
		fmt.Sprintln(&lastName)
		fmt.Println()

		fmt.Println("Email: ")
		fmt.Sprintln(&email)
		fmt.Println()
	} else {
		firstName = cl.Config.UserProfile.FirstName
		lastName = cl.Config.UserProfile.LastName
		email = cl.Config.UserProfile.Email
	}

	// Create the Files Needed By the Cluster

	projectSrcFolderPath := projectPath.Dir("src")

	// see if a cluster file with that Name already exists
	clusterPath := projectSrcFolderPath.File(clusterName + ".etl.go")
	if _, err := os.Stat(clusterPath.ToString()); err == nil {
		fmt.Println("a cluster with the Name of (" + clusterName + ") already exists")
		return true
	}

	clusterTemplatePath := TemplateFolder().File("Name.etl.go")
	if !clusterTemplatePath.Exists() {
		fmt.Println("cluster template file missing")
		return true
	}

	firstLetterOfClusterName := clusterName[:]

	// change the first letter in the cluster name to lower case
	var clusterNameCamelCase string
	idx := 0
	for len(firstLetterOfClusterName) > 0 {
		letter, size := utf8.DecodeRuneInString(firstLetterOfClusterName)
		if idx == 0 {
			letter = unicode.ToLower(letter)
		}
		clusterNameCamelCase = clusterNameCamelCase + string(letter)

		firstLetterOfClusterName = firstLetterOfClusterName[size:]
	}

	var processedTemplate []byte
	if bytes, err := clusterTemplatePath.Read(); err != nil {
		fmt.Println("cluster template file corrupted")
		return true
	} else {
		match := make(map[string]string)
		match["project"] = projectConfig.Name
		match["first-name"] = firstName
		match["last-name"] = lastName
		match["email"] = email
		match["cluster"] = clusterName
		match["cluster-short"] = clusterNameCamelCase

		// read the cluster template
		processedTemplate = template.Process(bytes, match)
	}

	// write the processed file

	clusterProjectPath := projectSrcFolderPath.File(clusterName + ".etl.go")
	os.Create(clusterProjectPath.ToString())
	if err := clusterProjectPath.Write(processedTemplate); err != nil {
		fmt.Println("failed to write cluster file")
		return true
	}

	// add the cluster to the root
	var processedRootGoFile string

	rootGoProjectPath := projectPath.File(projectConfig.Name + ".root.go")
	if bytes, err := rootGoProjectPath.Read(); err != nil {
		fmt.Println("missing project root go file")
	} else {
		stringRepOfBytes := string(bytes)

		var line string
		for c := range stringRepOfBytes {
			char := stringRepOfBytes[c]
			if char != '\n' {
				line += string(char)
			} else {
				if strings.Contains(line, "// DEFINED CLUSTERS END") {
					processedRootGoFile += "\t" + clusterNameCamelCase + " := " + clusterName + "{}\n"
					processedRootGoFile += "\tc.Cluster(\"" + clusterName + "\", " + clusterNameCamelCase + ", cluster.Config{Identifier: \"" + clusterName + "\"}))\n\n"
				}
				processedRootGoFile += line + "\n"

				line = ""
			}
		}
	}

	if _, err := os.Stat(rootGoProjectPath.ToString()); err != nil {
		fmt.Println("the project is missing a root folder, is this the wrong directory?")
		return true
	}

	// the file exists so it needs to be removed, otherwise do nothing
	if err := rootGoProjectPath.Remove(); err != nil {
		fmt.Println("failed to remove outdated root file")
		return true
	}

	os.Create(rootGoProjectPath.ToString())
	if err := rootGoProjectPath.Write([]byte(processedRootGoFile)); err != nil {
		fmt.Println("failed to write new root file")
		return true
	}

	// generate test file

	// does a stat file exist in the test folder with the provided Name?
	testFilePath := projectPath.Dir("test").File(clusterName + ".etl.test.go")
	if testFilePath.Exists() {
		fmt.Println("a test file already exists with the cluster Name (" + clusterName + ")")
		return true
	}

	// if a test file doesn't, write it to the
	var testFileBytes []byte
	testTemplateFilePath := TemplateFolder().File("Name.etl.test.go")
	if bytes, err := testTemplateFilePath.Read(); err != nil {
		fmt.Println("the test file template is missing")
		return true
	} else {
		match := make(map[string]string)

		testFileBytes = template.Process(bytes, match)
	}

	os.Create(testFilePath.ToString())
	if err := testFilePath.Write(testFileBytes); err != nil {
		fmt.Println("could not write test file for cluster")
		return true
	}

	// complete
	fmt.Println("Cluster " + clusterName + " was created")
	return true
}

func (cc ClusterCommand) DeleteCluster(cl *commandline.CommandLine) commandline.TerminateOnCompletion {

	// confirm that that use wants to permanently delete the source file
	clusterName := cl.NextArg()
	if clusterName == commandline.FinalArg {
		fmt.Println("(!) missing cluster name")
		return true
	}

	fmt.Print("are you sure you want to delete the cluster " + clusterName + "? [Y/n] ")
	var response string
	fmt.Scanln(&response)
	if (response != "Y") && (response != "") {
		return true
	}
	fmt.Println()

	srcFolder := commandline.EmptyPath().Dir(cl.MetaData.WorkingDirectory).Dir("src")
	testFolder := commandline.EmptyPath().Dir(cl.MetaData.WorkingDirectory).Dir("test")

	srcFile := srcFolder.File(clusterName + ".etl.go")
	if srcFile.Exists() {
		srcFile.Remove()
	}

	testFile := testFolder.File(clusterName + ".etl.test.go")
	if testFile.Exists() {
		testFile.Remove()
	}

	fmt.Println("Cluster " + clusterName + " was deleted")
	return true
}

func (cc ClusterCommand) ShowClusters(cl *commandline.CommandLine) commandline.TerminateOnCompletion {

	srcFolder := commandline.EmptyPath().Dir(cl.MetaData.WorkingDirectory).Dir("src")

	files, err := ioutil.ReadDir(srcFolder.ToString())
	if err != nil {
		return true
	}

	for _, fileInfo := range files {
		fmt.Print(fileInfo.Name()[:len(fileInfo.Name())-7]) // remove the ".etl.go" that is appended to the end of every file

		// read the contents of the file to get when it was created and by who
		file, err := os.Open(srcFolder.ToString() + fileInfo.Name())
		if err != nil {
			fmt.Println(err)
			continue
		}

		// in a file the "created on" should appear before the "created by" metadata
		// ! once we see the created by data we can ignore the rest of the file
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "Generated On") {
				split := strings.Split(scanner.Text(), " ")
				dateAndTime := split[len(split)-2:]
				fmt.Printf(" (Created on %s %s)", dateAndTime[0], dateAndTime[1])
			} else if strings.Contains(scanner.Text(), "Generated By") {
				split := strings.Split(scanner.Text(), " ")
				firstAndLastAndEmail := split[len(split)-3:]
				fmt.Printf(" (Created by %s %s %s)", firstAndLastAndEmail[0], firstAndLastAndEmail[1], firstAndLastAndEmail[2])

				break // we don't care about the contents in the rest of the file
			}
		}
		fmt.Println()

		file.Close()
	}

	return true
}

// DEVELOPER PROFILE COMMAND START

type ProfileCommand struct {
	name string
}

func (pc ProfileCommand) Name() string {
	return pc.name
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
