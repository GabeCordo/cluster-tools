package cli

import (
	"errors"
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/core"
	"github.com/GabeCordo/cluster-tools/internal/core/threads/common"
	"github.com/GabeCordo/commandline"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"strconv"
	"strings"
)

type ConfigCommand struct {
}

var FailedParsing = errors.New("failed to parse config field")
var BadValue = errors.New("the field/value does not match")

func (command ConfigCommand) configParser(c *core.Config, fields []string) (any, error) {

	numOfFields := len(fields)

	if fields[0] == "database" {

		if numOfFields < 2 {
			return nil, FailedParsing
		}

		if fields[1] == "type" {
			return &c.Database.Type, nil
		} else {
			return nil, FailedParsing
		}
	} else if fields[0] == "processor" {

		if numOfFields < 2 {
			return nil, FailedParsing
		}

		if fields[1] == "probe-every" {
			return &c.Processor.ProbeEvery, nil
		} else if fields[1] == "max-retries" {
			return &c.Processor.MaxRetry, nil
		} else {
			return nil, FailedParsing
		}
	} else if fields[0] == "mount-by-default" {
		return &c.MountByDefault, nil
	} else if fields[0] == "repl" {
		return &c.EnableRepl, nil
	} else if fields[0] == "debug" {
		return &c.Debug, nil
	} else {
		return nil, FailedParsing
	}
}

func (command ConfigCommand) printField(c *core.Config, fields []string) error {

	fieldRef, err := command.configParser(c, fields)
	if err != nil {
		return err
	}

	switch (fieldRef).(type) {
	case *string:
		{
			fmt.Println(*fieldRef.(*string))
		}
	case *int:
		{
			fmt.Println(*fieldRef.(*int))
		}
	case *uint32:
		{
			fmt.Println(*fieldRef.(*uint32))
		}
	case *float64:
		{
			fmt.Println(*fieldRef.(*float64))
		}
	case *bool:
		{
			fmt.Println(*fieldRef.(*bool))
		}
	}
	return nil
}

func (command ConfigCommand) updateField(c *core.Config, fields []string, value string) error {

	fieldRef, err := command.configParser(c, fields)
	if err != nil {
		return err
	}

	switch (fieldRef).(type) {
	case *string:
		{
			tmp := fieldRef.(*string)
			*tmp = value
		}
	case *int:
		{
			tmp := fieldRef.(*int)
			if i, err := strconv.Atoi(value); err != nil {
				return err
			} else {
				*tmp = i
			}
		}
	case *uint32:
		{
			tmp := fieldRef.(*uint32)
			if i, err := strconv.ParseUint(value, 10, 32); err != nil {
				return err
			} else {
				*tmp = uint32(i)
			}
		}
	case *float64:
		{
			tmp := fieldRef.(*float64)
			if i, err := strconv.ParseFloat(value, 64); err != nil {
				return err
			} else {
				*tmp = i
			}
		}
	case *bool:
		{
			tmp := fieldRef.(*bool)
			if i, err := strconv.ParseBool(value); err != nil {
				return err
			} else {
				*tmp = i
			}
		}
	}

	return nil
}

func (command ConfigCommand) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	if _, err := os.Stat(common.DefaultConfigFile); err != nil {
		fmt.Println("[x] cluster.tools has never been initialized, run 'cluster-tools init'")
		return commandline.Terminate
	}

	configFile, err := os.Open(common.DefaultConfigFile)
	if err != nil {
		fmt.Printf("[x] the global config file is missing (%s)\n", common.DefaultConfigFile)
		return commandline.Terminate
	}

	bytes, err := io.ReadAll(configFile)
	if err != nil {
		fmt.Printf("[x] the global config file is corrupt (%s)\n", common.DefaultConfigFile)
		return commandline.Terminate
	}

	configFile.Close()

	c := &core.Config{}
	if err := yaml.Unmarshal(bytes, c); err != nil {
		fmt.Printf("[x] the global config is corrupt (%s)\n", common.DefaultConfigFile)
		return commandline.Terminate
	}

	field := cli.NextArg()
	if field == commandline.FinalArg {
		fmt.Printf("[x] missing argument(1); the field in the config to modify\n")
		return commandline.Terminate
	}

	// allow the operator/developer to use dot-notation to reference a field
	parsedFields := strings.Split(field, ".")

	if cli.Flag(commandline.Show) {

		if err := command.printField(c, parsedFields); err != nil {
			fmt.Printf("[x] %s\n", err)
		}

	} else if cli.Flag(commandline.Update) {

		value := cli.NextArg()
		if value == commandline.FinalArg {
			fmt.Printf("[x] missing argument(2); the value in the config to modify\n")
			return commandline.Terminate
		}

		if err := command.updateField(c, parsedFields, value); err != nil {
			fmt.Printf("[x] %s\n", err)
			return commandline.Terminate
		}

		updatedBytes, err := yaml.Marshal(c)
		if err != nil {
			fmt.Printf("[Error] failed to marshal updated config => %s\n", err.Error())
			return commandline.Terminate
		}

		if configFile, err = os.Create(common.DefaultConfigFile); err != nil {
			fmt.Printf("[Error] failed to truncate old config file => %s\n", err.Error())
			return commandline.Terminate
		}
		defer configFile.Close()

		if _, err = configFile.Write(updatedBytes); err != nil {
			fmt.Printf("[Error] failed to write bytes to config file => %s\n", err.Error())
		}
	}

	return commandline.Terminate
}
