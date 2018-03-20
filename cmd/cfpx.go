package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
)

type keymap map[string]interface{}
type delimiterMap map[string]keymap

const (
	defaultMappingFilename = "argmap.toml"
	commandKey             = "command"
	commandNameKey         = "name"
	argsKey                = "args"
	proxySuffix            = "-proxy"
)

const (
	_ = iota // Reserve 0 for success
	errorMissingMappingFIle
	errorInvalidCommandNameType
	errorCommandInvocationError
	errorNoCommandNameProvided
)

var filename string

func init() {
	flag.StringVar(&filename, "m", defaultMappingFilename, "Path to the TOML-formatted argument mapping file.")
}

func findMappingFile(filepath string) (string, error) {
	if !path.IsAbs(filepath) {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		filepath = path.Join(wd, filepath)
	}
	if _, err := os.Stat(filepath); err != nil {
		return "", err
	}
	return filepath, nil
}

func getExecutableName(invokedName, commandName string) string {
	if commandName != "" {
		return commandName
	}
	if strings.HasSuffix(invokedName, proxySuffix) {
		return strings.TrimSuffix(invokedName, proxySuffix)
	}
	return ""
}

func extractCommandName(commandName interface{}) (string, error) {
	if commandName == nil {
		return "", nil
	}
	commandString, ok := commandName.(string)
	if !ok {
		return "", errors.New("cfpx: command.name must be a string")
	}
	return commandString, nil
}

func extractStringArgs(rawArgs interface{}) []string {
	var args []string
	if rawArgs != nil && reflect.TypeOf(rawArgs).Kind() == reflect.Slice {
		argSlice := rawArgs.([]interface{})
		args = make([]string, 0, len(argSlice))
		for _, val := range argSlice {
			if reflect.TypeOf(val).Kind() == reflect.String {
				args = append(args, reflect.ValueOf(val).String())
			}
		}
	}
	return args
}

func main() {
	flag.Parse()

	filepath, err := findMappingFile(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(errorMissingMappingFIle)
	}

	mapping := make(delimiterMap)
	_, err = toml.DecodeFile(filepath, &mapping)
	if err != nil {
		fmt.Println(err)
	}

	commandName, err := extractCommandName(mapping[commandKey][commandNameKey])
	if err != nil {
		fmt.Println(err)
		os.Exit(errorInvalidCommandNameType)
	}

	executableName := getExecutableName(os.Args[0], commandName)
	if executableName == "" {
		fmt.Println("cfpx: no command name provided")
		os.Exit(errorNoCommandNameProvided)
	}

	cmd := exec.Command(executableName, extractStringArgs(mapping[commandKey][argsKey])...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		fmt.Println(err)
		os.Exit(errorCommandInvocationError)
	}
}
