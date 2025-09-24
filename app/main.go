package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var COMMANDS map[string]func([]string)

func init() {
	COMMANDS = map[string]func([]string){
		"exit": exit,
		"echo": echo,
		"type": whatType,
		"pwd":  pwd,
	}
}

func main() {
	for {
		fmt.Fprint(os.Stdout, "$ ")
		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return
		}
		evaluate(input)
	}
}

func evaluate(rawInput string) {
	input := strings.TrimSpace(rawInput)
	if input == "" {
		return
	}

	args := strings.Split(input, " ")
	command, optional := args[0], args[1:]

	output, ok := COMMANDS[command]

	if ok {
		output(optional)
	} else {
		fmt.Println(command + ": command not found")
	}
}

func exit(input []string) {
	if strings.TrimSpace(input[len(input)-1]) != "0" {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func echo(input []string) {
	if len(input) < 1 {
		fmt.Println("error: missing operand for echo.")
	}
	fmt.Print(strings.Join(input, " "))
}

func whatType(input []string) {
	paths := strings.Split(os.Getenv("PATH"), ":")
	if len(input) < 1 {
		fmt.Println("error: missing operand for type.")
	}

	ourCommand := input[0][:len(input[0])-1]

	_, ok := COMMANDS[ourCommand]
	if ok {
		fmt.Printf("%v is a shell builtin\n", ourCommand)
	} else {
		filePath := findExecutable(ourCommand, paths)

		if filePath != "" {
			fmt.Printf("%s is %s\n", ourCommand, filePath)
		} else {

			fmt.Println(ourCommand + ": command not found")
		}
	}
}

func findExecutable(command string, paths []string) string {
	for _, path := range paths {
		filePath := filepath.Join(path, command)
		fileInfo, err := os.Stat(filePath)
		if err == nil && fileInfo.Mode().Perm()&0111 != 0 {
			return filePath
		}
	}

	return ""
}

func pwd(input []string) {
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fmt.Println(currentWorkingDirectory)
}
