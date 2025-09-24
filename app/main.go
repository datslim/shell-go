package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
)

var COMMANDS map[string]func([]string)

func init() {
	COMMANDS = map[string]func([]string){
		"exit":    exit,
		"echo":    echo,
		"type":    whatType,
		"pwd":     pwd,
		"cd":      cd,
		"ls":      ls,
		"history": history,
		"clear":   clear,
	}
}

func main() {
	autoCompleter := readline.NewPrefixCompleter(
		readline.PcItem("exit"),
		readline.PcItem("echo"),
		readline.PcItem("type"),
		readline.PcItem("pwd"),
		readline.PcItem("cd"),
		readline.PcItem("ls"),
		readline.PcItem("history"),
	)

	l, err := readline.NewEx(&readline.Config{
		AutoComplete: autoCompleter,
		HistoryFile:  "/tmp/shell_history",
	})

	if err != nil {
		log.Fatal(err)
	}

	defer l.Close()

	for {
		currDir, _ := os.Getwd()
		homeDir, _ := os.UserHomeDir()
		var beautifulPwd string
		if after, found := strings.CutPrefix(currDir, homeDir); found {
			beautifulPwd = "~" + after
		} else {
			beautifulPwd = currDir
		}

		prompt := color.CyanString("%s ", beautifulPwd) + "$ "
		l.SetPrompt(prompt)

		input, err := l.Readline()
		if err != nil {
			return
		}

		execute(input)
	}
}

func execute(rawInput string) {
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
		color.Red(command + ": command not found")
	}
}

func exit(input []string) {
	if len(input) == 0 {
		os.Exit(0)
	}

	if strings.TrimSpace(input[len(input)-1]) != "0" {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func echo(input []string) {
	if len(input) < 1 {
		color.Red("error: missing operand for echo.")
		return
	}
	fmt.Println(strings.Join(input, " "))
}

func whatType(input []string) {
	paths := strings.Split(os.Getenv("PATH"), ":")
	if len(input) < 1 {
		color.Red("error: missing operand for type.")
		return
	}

	ourCommand := input[0]

	_, ok := COMMANDS[ourCommand]
	if ok {

		fmt.Printf("%v is a shell builtin\n", color.GreenString("%v", ourCommand))
	} else {
		filePath := findExecutable(ourCommand, paths)

		if filePath != "" {
			fmt.Printf("%s is %s\n", ourCommand, filePath)
		} else {
			color.Red(ourCommand + ": command not found")
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

func cd(directory []string) {
	homeDirectory := os.Getenv("HOME")
	if len(directory) == 0 {
		os.Chdir(homeDirectory)
	} else if strings.HasPrefix(directory[0], "/") || strings.HasPrefix(directory[0], "~") {
		toAbsolutePath(homeDirectory, directory)
	} else {
		toRelativePath(directory[0])
	}
}

func toAbsolutePath(homeDirectory string, directory []string) {
	destination := directory[len(directory)-1]
	destination = strings.ReplaceAll(destination, "~", homeDirectory)
	err := os.Chdir(destination)
	if err != nil {
		fmt.Printf("cd: %v:", destination)
	}
}

func toRelativePath(directory string) {
	currentDirectory, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	resultPath := filepath.Join(currentDirectory, directory)
	err = os.Chdir(resultPath)
	if err != nil {
		color.Red("cd: %v: No such file or directory\n", directory)
	}
}

func ls(input []string) {
	currentDirectory, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	files, err := os.ReadDir(currentDirectory)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		color.HiMagenta(file.Name())
	}
}

func history(input []string) {
	file, err := os.Open("/tmp/shell_history")
	if err != nil {
		return
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if len(input) > 0 {
		n, err := strconv.Atoi(input[0])
		if err != nil {
			color.Red("error: argument should be a number.")
			return
		}
		start := len(lines) - n
		if start < 0 {
			start = 0
		}
		lines = lines[start:]
	}

	for i, line := range lines {
		color.Set(color.FgHiMagenta)
		fmt.Printf(" %d  ", i+1)
		color.Unset()
		fmt.Printf("%s\n", line)
	}
}

func clear(input []string) {
	readline.ClearScreen(os.Stdout)
}
