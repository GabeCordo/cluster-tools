package cli

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"etl/core"
	"etl/net"
	"etl/utils/template"
	"fmt"
	"io/ioutil"
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

// DEBUG COMMAND START

type DebugCommand struct {
	name string
}

func (dc DebugCommand) Name() string {
	return dc.name
}

func (dc DebugCommand) Run(cli *CommandLine) Terminate {
	cli.Flags.Debug = true

	return false // do not terminate
}

// CREATE COMMAND START

type CreateCommand struct {
	name string
}

func (cc CreateCommand) Name() string {
	return cc.name
}

func (cc CreateCommand) Run(cli *CommandLine) Terminate {
	cli.Flags.Create = true

	return false // do not terminate
}

// DELETE COMMAND START

type DeleteCommand struct {
	name string
}

func (dc DeleteCommand) Name() string {
	return dc.name
}

func (dc DeleteCommand) Run(cli *CommandLine) Terminate {
	cli.Flags.Delete = true

	return false // do not terminate
}

// SHOW COMMAND START

type ShowCommand struct {
	name string
}

func (sc ShowCommand) Name() string {
	return sc.name
}

func (sc ShowCommand) Run(cli *CommandLine) Terminate {
	cli.Flags.Show = true

	return false
}

// VERSION COMMAND START

type VersionCommand struct {
	name string
}

func (vc VersionCommand) Name() string {
	return vc.name
}

func (vc VersionCommand) Run(cli *CommandLine) Terminate {
	strVersion := fmt.Sprintf("%.2f", cli.Config.Version)
	strTimeNow := time.Now().Format("Mon Jan _2 15:04:05 MST 2006")
	fmt.Println("ETLFramework Version " + strVersion + " " + strTimeNow)
	return true
}

// HELP COMMAND START

type HelpCommand struct {
	name string
}

func (helpCommand HelpCommand) Name() string {
	return helpCommand.name
}

func (helpCommand HelpCommand) Run(cli *CommandLine) Terminate {
	fmt.Println("etl")
	fmt.Println("-h\tView helpful information about the etl service")
	fmt.Println("-d\tEnable debug mode")
	fmt.Println("-g\tGenerate an ECDSA x509 public and private key pair")

	return true
}

// GENERATE KEY PAIR START

type GenerateKeyPairCommand struct {
	name string
}

func (gkpc GenerateKeyPairCommand) Name() string {
	return gkpc.name
}

func (gkpc GenerateKeyPairCommand) Run(cli *CommandLine) Terminate {

	// we only want to create a key if it has been pre-pended with the 'create' flag
	if cli.Flags.Create {
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

func (idc InteractiveDashboardCommand) Run(cli *CommandLine) Terminate {
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

func (dc *DeployCommand) Run(cli *CommandLine) Terminate {

	if cli.Flags.Debug {
		log.Println("(+) starting up etl")
	}

	cli.Core.Run()

	if cli.Flags.Debug {
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

func (cproj CreateProjectCommand) Run(cli *CommandLine) Terminate {

	if !cli.Flags.Create {
		fmt.Println("missing create flag")
		return true
	}

	var projectName string
	if projectName = cli.NextArg(); projectName == FinalArg {
		fmt.Println("missing project name")
		return true
	}

	projectPath := cli.MetaData.WorkingDirectory + "/" + projectName

	if err := os.Mkdir(projectPath, DefaultFilePermissions); err != nil {
		fmt.Println("failed to create project directory")
		return true
	}

	projectConfig := core.NewConfig(projectName)

	projectConfigPath := projectPath + "/config.etl.json"
	projectConfig.ToJson(projectConfigPath)

	// GENERATE THE DEFAULT PROJECT FILES
	executablePath, _ := os.Executable()

	if executablePath[1:8] == "private" {
		executablePath = "/Users/gabecordovado/go/src/etl/"
	} else {
		executablePath = executablePath[:len(executablePath)-9]
	}
	templateFolderPath := executablePath + ".bin/templates/"

	stringRepOfModFile := "module " + projectName + "\n\ngo " + runtime.Version()[2:]
	if err := ioutil.WriteFile(projectPath+"/go.mod", []byte(stringRepOfModFile), DefaultFilePermissions); err != nil {
		fmt.Println("failed to create go module")
		return true
	}

	var processedRootTemplateFile []byte
	projectRootTemplatePath := templateFolderPath + "project.root.go"
	if bytes, err := ioutil.ReadFile(projectRootTemplatePath); err == nil {
		match := make(map[string]string)
		match["project"] = projectName
		processedRootTemplateFile = template.Process(bytes, match)
	} else {
		fmt.Println(err)
		fmt.Println("a template file is missing")
	}

	rootGoFilePath := projectPath + "/" + projectName + ".root.go"
	if err := ioutil.WriteFile(rootGoFilePath, processedRootTemplateFile, DefaultFilePermissions); err != nil {
		fmt.Println("failed to create root project go file")
		return true
	}

	if err := os.Mkdir(projectPath+"/src", DefaultFilePermissions); err != nil {
		fmt.Println("failed to create project files")
		return true
	}

	vectorEtlTemplatePath := templateFolderPath + "vector.etl.go"
	fmt.Println(vectorEtlTemplatePath)
	if exampleVectorEtlBytes, err := ioutil.ReadFile(vectorEtlTemplatePath); err == nil {

		vectorEtlProjectPath := projectPath + "/src/vector.etl.go"
		if err = ioutil.WriteFile(vectorEtlProjectPath, exampleVectorEtlBytes, DefaultFilePermissions); err != nil {
			fmt.Println("failed writing default Vector example to project")
			return true
		}
	} else {
		fmt.Println("failed reading default Vector example from templates")
		return true
	}

	if err := os.Mkdir(projectPath+"/test", DefaultFilePermissions); err != nil {
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

func (ccc CreateClusterCommand) Run(cli *CommandLine) Terminate {
	projectPath := cli.MetaData.WorkingDirectory + "/"

	// see if we are currently in an etl project, otherwise we cannot add a cluster
	configPath := projectPath + "config.etl.json"
	if _, err := os.Stat(configPath); err != nil {
		fmt.Println("no etl project exists")
		return true
	}

	// read the config file
	var projectConfig core.Config
	core.JSONToETLConfig(&projectConfig, configPath)

	// the cluster name should be stored as the second argument
	clusterName := cli.NextArg()
	if clusterName == FinalArg {
		fmt.Println("missing cluster name")
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

	projectSrcFolderPath := projectPath + "/src/"

	// see if a cluster file with that name already exists
	clusterPath := projectSrcFolderPath + clusterName + ".etl.go"
	if _, err := os.Stat(clusterPath); err == nil {
		fmt.Println("a cluster with the name of (" + clusterName + ") already exists")
		return true
	}

	executablePath, _ := os.Executable()
	if executablePath[1:8] == "private" {
		executablePath = "/Users/gabecordovado/go/src/etl/"
	} else {
		executablePath = executablePath[:len(executablePath)-9]
	}
	templateFolderPath := executablePath + ".bin/templates/"

	clusterTemplatePath := templateFolderPath + "name.etl.go"
	if _, err := os.Stat(clusterTemplatePath); err != nil {
		fmt.Println("cluster template file missing")
		return true
	}

	firstLetterOfClusterName := clusterName[:1]
	unicodeFirstLetterOfClusterName, _ := utf8.DecodeRuneInString(firstLetterOfClusterName)
	firstLetterOfClusterName = string(unicode.ToLower(unicodeFirstLetterOfClusterName))

	var proccessedTemplate []byte
	if bytes, err := ioutil.ReadFile(clusterTemplatePath); err != nil {
		fmt.Println("cluster template file corrupted")
		return true
	} else {
		match := make(map[string]string)
		match["project"] = projectConfig.Name
		match["first-name"] = cli.Config.UserProfile.FirstName
		match["last-name"] = cli.Config.UserProfile.LastName
		match["email"] = cli.Config.UserProfile.Email
		match["cluster"] = clusterName
		match["cluster-short"] = firstLetterOfClusterName

		// read the cluster template
		proccessedTemplate = template.Process(bytes, match)
	}

	// write the processed file

	clusterProjectPath := projectSrcFolderPath + clusterName + ".etl.go"
	if err := ioutil.WriteFile(clusterProjectPath, proccessedTemplate, DefaultFilePermissions); err != nil {
		fmt.Println("failed to write cluster file")
		return true
	}

	// add the cluster to the root
	var processedRootGoFile string

	rootGoProjectPath := projectPath + projectConfig.Name + ".root.go"
	if bytes, err := ioutil.ReadFile(rootGoProjectPath); err != nil {
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
	if err := os.Remove(rootGoProjectPath); err != nil {
		fmt.Println("failed to remove outdated root file")
		return true
	}

	if err := ioutil.WriteFile(rootGoProjectPath, []byte(processedRootGoFile), DefaultFilePermissions); err != nil {
		fmt.Println("failed to write new root file")
		return true
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

func (pc ProfileCommand) Run(cli *CommandLine) Terminate {

	if cli.Flags.Create {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("First Name: ")
		line, err := reader.ReadString('\n')
		if err != nil {
			panic("error reading profile")
		}
		cli.Config.UserProfile.FirstName = line[:len(line)-1] // remove the delim

		fmt.Print("Last Name: ")
		line, err = reader.ReadString('\n')
		if err != nil {
			panic("error reading profile")
		}
		cli.Config.UserProfile.LastName = line[:len(line)-1] // remove the delim

		fmt.Print("Email: ")
		line, err = reader.ReadString('\n')
		if err != nil {
			panic("error reading profile")
		}
		cli.Config.UserProfile.Email = line[:len(line)-1] // remove the delim

		cli.Config.ToJson() // push the JSON update to the local file
	} else if cli.Flags.Show {
		if (len(cli.Config.UserProfile.FirstName) == 0) && (len(cli.Config.UserProfile.LastName) == 0) && (len(cli.Config.UserProfile.Email) == 0) {
			fmt.Println("developer profile not configured")
			fmt.Println("use \"etl create profile\" to create a new developer profile")
		} else {
			strVersion := fmt.Sprintf("%.2f", cli.Config.Version)
			fmt.Println("ETLFramework [Version " + strVersion + "]")
			fmt.Println(cli.Config.UserProfile.FirstName + " " + cli.Config.UserProfile.LastName)
			fmt.Println(cli.Config.UserProfile.Email)
		}
	} else {
		fmt.Println("missing flag directive")
	}

	return true
}
