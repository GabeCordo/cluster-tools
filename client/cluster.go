package client

import (
	"bufio"
	"fmt"
	"github.com/GabeCordo/commandline"
	"github.com/GabeCordo/etl/core"
	"github.com/GabeCordo/etl/utils/template"
	"io/ioutil"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)

// CLUSTER COMMAND START

type ClusterCommand struct {
	PublicName string
}

func (cc ClusterCommand) Name() string {
	return cc.PublicName
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

	// change the first letter in the cluster PublicName to lower case
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
		match["first-PublicName"] = firstName
		match["last-PublicName"] = lastName
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
		fmt.Println("(!) missing cluster PublicName")
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
