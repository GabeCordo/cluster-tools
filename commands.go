package main

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"etl/core"
	"etl/net"
	"etl/utils"
	"etl/utils/cli"
	"etl/utils/template"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unicode"
	"unicode/utf8"
)

// VERSION COMMAND START

type VersionCommand struct {
	name string
}

func (vc VersionCommand) Name() string {
	return vc.name
}

func (vc VersionCommand) Run(cl *cli.CommandLine) cli.Terminate {
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

func (gkpc GenerateKeyPairCommand) Run(cl *cli.CommandLine) cli.Terminate {

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
		fmt.Println(net.ByteToString(x509Encoded))

		x509EncodedPub, err := x509.MarshalPKIXPublicKey(&pair.PublicKey)
		fmt.Println(len(x509EncodedPub))
		fmt.Println("[public]")
		fmt.Println(net.ByteToString(x509EncodedPub))
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

func (idc InteractiveDashboardCommand) Run(cl *cli.CommandLine) cli.Terminate {
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

func (dc DeployCommand) Run(cl *cli.CommandLine) cli.Terminate {

	if cl.Flags.Debug {
		log.Println("(+) starting up etl")
	}

	cl.Core.Run()

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

func (cproj CreateProjectCommand) Run(cl *cli.CommandLine) cli.Terminate {

	if !cl.Flags.Create {
		fmt.Println("missing create flag")
		return true
	}

	var projectName string
	if projectName = cl.NextArg(); projectName == cli.FinalArg {
		fmt.Println("missing project Name")
		return true
	}

	projectPath := utils.EmptyPath().Dir(cl.MetaData.WorkingDirectory).Dir(projectName)

	if err := projectPath.MkDir(); err != nil {
		fmt.Println("failed to create project directory")
		return true
	}

	projectConfig := core.NewConfig(projectName)

	projectConfigPath := projectPath.File("/config.etl.json")
	projectConfig.ToJson(projectConfigPath.ToString())

	// GENERATE THE DEFAULT PROJECT FILES
	templateFolderPath := cli.TemplateFolderPath()

	stringRepOfModFile := "module " + projectName + "\n\ngo " + runtime.Version()[2:]

	projectGoModulePath := projectPath.File("go.mod")
	if err := projectGoModulePath.Write([]byte(stringRepOfModFile)); err != nil {
		fmt.Println("failed to create go module")
		return true
	}

	var processedRootTemplateFile []byte
	projectRootTemplatePath := templateFolderPath.File("project.root.go")
	if bytes, err := projectRootTemplatePath.Read(); err == nil {
		match := make(map[string]string)
		match["project"] = projectName
		processedRootTemplateFile = template.Process(bytes, match)
	} else {
		fmt.Println(err)
		fmt.Println("a template file is missing")
	}

	rootGoFilePath := projectPath.File(projectName + ".root.go")
	if err := rootGoFilePath.Write(processedRootTemplateFile); err != nil {
		fmt.Println("failed to create root project go file")
		return true
	}

	projectSrcDirPath := projectPath.Dir("src")
	if err := projectSrcDirPath.MkDir(); err != nil {
		fmt.Println("failed to create project files")
		return true
	}

	vectorEtlTemplatePath := templateFolderPath.File("vector.etl.go")
	if exampleVectorEtlBytes, err := vectorEtlTemplatePath.Read(); err == nil {

		vectorEtlProjectPath := projectPath.Dir("src").File("vector.etl.go")
		if err = vectorEtlProjectPath.Write(exampleVectorEtlBytes); err != nil {
			fmt.Println("failed writing default Vector example to project")
			return true
		}
	} else {
		fmt.Println("failed reading default Vector example from .templates")
		return true
	}

	projectTestDirectoryPath := projectPath.Dir("test")
	if err := projectTestDirectoryPath.MkDir(); err != nil {
		fmt.Println("failed to create project files")
		return true
	}

	return true
}

// CREATE CLUSTER COMMAND START

type CreateClusterCommand struct {
	name string
}

func (ccc CreateClusterCommand) Name() string {
	return ccc.name
}

func (ccc CreateClusterCommand) Run(cl *cli.CommandLine) cli.Terminate {
	projectPath := utils.EmptyPath().Dir(cl.MetaData.WorkingDirectory)

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
	if clusterName == cli.FinalArg {
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

	projectSrcFolderPath := projectPath.Dir("src")

	// see if a cluster file with that Name already exists
	clusterPath := projectSrcFolderPath.File(clusterName + ".etl.go")
	if _, err := os.Stat(clusterPath.ToString()); err == nil {
		fmt.Println("a cluster with the Name of (" + clusterName + ") already exists")
		return true
	}

	templateFolderPath := cli.TemplateFolderPath()

	clusterTemplatePath := templateFolderPath.File("Name.etl.go")
	if !clusterTemplatePath.Exists() {
		fmt.Println("cluster template file missing")
		return true
	}

	firstLetterOfClusterName := clusterName[:1]
	unicodeFirstLetterOfClusterName, _ := utf8.DecodeRuneInString(firstLetterOfClusterName)
	firstLetterOfClusterName = string(unicode.ToLower(unicodeFirstLetterOfClusterName))

	var processedTemplate []byte
	if bytes, err := clusterTemplatePath.Read(); err != nil {
		fmt.Println("cluster template file corrupted")
		return true
	} else {
		match := make(map[string]string)
		match["project"] = projectConfig.Name
		match["first-Name"] = cl.Config.UserProfile.FirstName
		match["last-Name"] = cl.Config.UserProfile.LastName
		match["email"] = cl.Config.UserProfile.Email
		match["cluster"] = clusterName
		match["cluster-short"] = firstLetterOfClusterName

		// read the cluster template
		processedTemplate = template.Process(bytes, match)
	}

	// write the processed file

	clusterProjectPath := projectSrcFolderPath.File(clusterName + ".etl.go")
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
					processedRootGoFile += "\t" + firstLetterOfClusterName + " := " + clusterName + "{}\n"
					processedRootGoFile += "\tc.Cluster(\"" + clusterName + "\", " + firstLetterOfClusterName + ", cluster.Config{Identifier: \"" + clusterName + "\"}))\n\n"
				}
				processedRootGoFile += line + "\n"

				line = ""
			}
		}
	}

	fmt.Println(rootGoProjectPath)
	if err := rootGoProjectPath.Remove(); err != nil {
		fmt.Println("failed to remove outdated root file")
		return true
	}

	if err := rootGoProjectPath.Write([]byte(processedRootGoFile)); err != nil {
		fmt.Println("failed to write new root file")
		return true
	}

	// generate test file

	// does a stat file exist in the test folder with the provided Name?
	testFilePath := projectPath.Dir("test").File(clusterName + "etl.test.go")
	if !testFilePath.Exists() {
		fmt.Println("a test file already exists with the cluster Name (" + clusterName + ")")
		return true
	}

	// if a test file doesn't, write it to the
	var testFileBytes []byte
	testTemplateFilePath := templateFolderPath.File("Name.etl.test.go")
	if bytes, err := testTemplateFilePath.Read(); err != nil {
		fmt.Println("the test file template is missing")
		return true
	} else {
		match := make(map[string]string)

		testFileBytes = template.Process(bytes, match)
	}

	if err := testFilePath.Write(testFileBytes); err != nil {
		fmt.Println("could not write test file for cluster")
		return true
	}

	// complete

	return true
}

// DEVELOPER PROFILE COMMAND START

type ProfileCommand struct {
	name string
}

func (pc ProfileCommand) Name() string {
	return pc.name
}

func (pc ProfileCommand) Run(cl *cli.CommandLine) cli.Terminate {

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

		cl.Config.ToJson() // push the JSON update to the local file
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
