package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
)

var COMMANDS map[string]func([]string)
var HISTORY = make([]string, 0)

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
	var historyfile string
	if histfile, exists := os.LookupEnv("HISTFILE"); exists {
		historyfile = histfile
	} else {
		historyfile = "/tmp/shell_history"
	}
	readHistoryFromFile(historyfile)

	autoCompleter := readline.NewPrefixCompleter(
		getExecutablesFromPATH()...,
	)
	fmt.Print(historyfile)
	l, err := readline.NewEx(&readline.Config{
		AutoComplete: autoCompleter,
		HistoryFile:  historyfile,
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
		if strings.TrimSpace(input) != "" {
			HISTORY = append(HISTORY, input)
			l.SaveHistory(input)
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
	paths := strings.Split(os.Getenv("PATH"), ":")
	execPath := findExecutable(command, paths)
	if ok {
		output(optional)
	} else if execPath != "" {
		cmd := exec.Command(execPath, optional...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		err := cmd.Run()
		if err != nil {
			color.Red("error: %v", err)
		}
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

func getExecutablesFromPATH() []readline.PrefixCompleterInterface {
	var items []readline.PrefixCompleterInterface

	for cmd := range COMMANDS {
		items = append(items, readline.PcItem(cmd))
	}

	paths := strings.Split(os.Getenv("PATH"), ":")
	seen := make(map[string]bool)

	for _, path := range paths {
		files, err := os.ReadDir(path)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			name := file.Name()
			if seen[name] {
				continue
			}

			filePath := filepath.Join(path, name)
			fileInfo, err := os.Stat(filePath)
			if err == nil && fileInfo.Mode().Perm()&0111 != 0 {
				items = append(items, readline.PcItem(name))
				seen[name] = true
			}
		}
	}

	return items
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
	if len(input) >= 2 {
		flag := input[0]
		path := input[1]

		switch flag {
		case "-r":
			readHistoryFromFile(path)
			return
		case "-a":
			appendHistoryToFile(path)
			return
		case "-w":
			appendHistoryToFile(path)
			return
		}
	}

	if len(input) == 1 && input[0] == "-r" {
		readHistoryFromFile("/tmp/shell_history")
		return
	}

	if len(input) == 1 {
		numberOfCommand, err := strconv.Atoi(input[0])
		if err != nil {
			color.Red("error: argument should be a number.")
			return
		}
		if numberOfCommand > len(HISTORY) {
			numberOfCommand = len(HISTORY)
		}
		start := len(HISTORY) - numberOfCommand
		for i := start; i < len(HISTORY); i++ {
			color.Set(color.FgHiMagenta)
			fmt.Printf(" %d  ", i+1)
			color.Unset()
			fmt.Printf("%s\n", HISTORY[i])
		}
		return
	}

	for i, command := range HISTORY {
		color.Set(color.FgHiMagenta)
		fmt.Printf(" %d  ", i+1)
		color.Unset()
		fmt.Printf("%s\n", command)
	}
}

func readHistoryFromFile(pathToHistory string) {
	_, err := os.Stat(pathToHistory)
	if err != nil {
		color.Red("error: history file not found.")
		return
	} else {
		file, err := os.Open(pathToHistory)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			HISTORY = append(HISTORY, scanner.Text())
		}

	}
}

func appendHistoryToFile(pathToHistory string) {
	file, err := os.OpenFile(pathToHistory, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	for _, command := range HISTORY {
		_, err := writer.WriteString(command + "\n")
		if err != nil {
			panic(err)
		}
	}
	writer.WriteString("\n")
	writer.Flush()
}

func clear(input []string) {
	readline.ClearScreen(os.Stdout)
}
