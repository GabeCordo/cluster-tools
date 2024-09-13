package controllers

import (
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/database/job"
	"github.com/GabeCordo/commandline"
	"gopkg.in/yaml.v3"
	"os"
	"strconv"
	"strings"
)

type ScheduleController struct {
}

// ParseTime
// turns an interval representation '-' or '-/n' into a duration gap.
func (controller ScheduleController) ParseTime(input string) int {

	if input == "-" {
		return 60
	}

	split := strings.Split(input, "/")
	if len(split) != 2 {
		return 60
	}

	if n, err := strconv.Atoi(split[1]); err != nil {
		return 60
	} else {
		return n
	}
}

func (controller ScheduleController) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	// PARSE THE INCOMING CLI ARGUMENTS
	jb := job.Job{}

	// the Identifier field is required by both the CREATE and DELETE flags
	jb.Identifier = cli.NextArg()
	if jb.Identifier == commandline.FinalArg {
		fmt.Println("missing the identifier parameter")
		return commandline.Terminate
	}

	// the module identifier is required by both the CREATE and DELETE flags
	jb.Module = cli.NextArg()
	if jb.Module == commandline.FinalArg {
		fmt.Println("missing the module parameter")
		return commandline.Terminate
	}

	// the Config and Minute identifiers are only required when creating a new job
	// asking for this information for a DELETE operation would be unnecessary.
	if cli.Flag(commandline.Create) {

		jb.Cluster = cli.NextArg()
		if jb.Cluster == commandline.FinalArg {
			fmt.Println("missing the cluster parameter")
			return commandline.Terminate
		}

		jb.Config = cli.NextArg()
		if jb.Config == commandline.FinalArg {
			fmt.Println("missing the config parameter")
			return commandline.Terminate
		}

		minutes := cli.NextArg()
		if minutes == commandline.FinalArg {
			fmt.Println("missing the minutes parameter")
			return commandline.Terminate
		} else {
			jb.Interval.Minute = controller.ParseTime(minutes)
		}
	}

	dump := &job.Dump{}

	filePath := fmt.Sprintf("%s/%s.yml", DefaultSchedulesFolder, jb.Module)
	if _, err := os.Stat(filePath); os.IsNotExist(err) && cli.Flag(commandline.Delete) {
		fmt.Printf("cannot delete scheduler file that does not exist(%s/%s.yml)\n",
			DefaultSchedulesFolder, jb.Module)
		return commandline.Terminate
	} else if err == nil {
		fmt.Println("[-] scheduler file exists ... pulling data")
		if b, err := os.ReadFile(filePath); err == nil {
			if err = yaml.Unmarshal(b, dump); err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Println(err)
		}
	}

	// DETERMINE IF THE MODULE FILE EXISTS IN THE SCHEDULER FOLDER

	if cli.Flag(commandline.Create) {

		if dump.Jobs == nil {
			dump.Jobs = make([]job.Job, 0)
		}
		dump.Jobs = append(dump.Jobs, jb)

		b, err := yaml.Marshal(dump)
		if err != nil {
			return commandline.Terminate
		}

		err = os.WriteFile(filePath, b, 0750)
		if err != nil {
			fmt.Println(err)
			return commandline.Terminate
		}

		fmt.Println("[-] added new job to static scheduler file successfully")

	} else {

		modifiedJobsList := make([]job.Job, 0)
		for _, j := range dump.Jobs {
			if j.Identifier != jb.Identifier {
				modifiedJobsList = append(modifiedJobsList, j)
			}
		}

		fmt.Println("[-] removed job from static scheduler successfully")

		if err := os.Remove(filePath); err != nil {
			return commandline.Terminate
		}

		if len(modifiedJobsList) > 0 {
			fmt.Println("[-] saving changes to static schedules file ...")

			b, err := yaml.Marshal(modifiedJobsList)
			if err != nil {
				fmt.Println(err)
				return commandline.Terminate
			}

			err = os.WriteFile(filePath, b, 0750)

			fmt.Println("[-] updated static scheduler successfully")
		}
	}

	return commandline.Terminate
}
