package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	ConfigFile        = "/.config/xmv"
	Timeout           = 10 * time.Second
	RunAllPatternsCmd = "zmv -n '%s' '%s'"
	RunOnePatternsCmd = "zmv '%s' '%s'"
)

type Config struct {
	Patterns map[string]string
}

type result struct {
	oldPattern string
	newPattern string
	samples    []string
}

var (
	config = Config{
		Patterns: make(map[string]string),
	}
	configFile = os.Getenv("HOME") + ConfigFile
)

func main() {
	f, err := os.Open(configFile)
	checkError(err, 1)

	err = json.NewDecoder(f).Decode(&config)
	checkError(err, 1)

	if len(os.Args) == 4 {
		oldPattern := os.Args[1]
		newPattern := os.Args[2]
		config.Patterns[oldPattern] = newPattern

		saveConfig()
	}

	succesfulPatterns := runAllPatterns()
	pattern := choosePattern(succesfulPatterns)
	runOnePattern(pattern)
}

func runZSH(command string) (string, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	exitCode := 0
	zshPath := findZSH()
	cmd := exec.CommandContext(ctx, zshPath, "-ic", command)
	output, err := cmd.Output()

	if exiterr, ok := err.(*exec.ExitError); ok {
		exitCode = exiterr.ExitCode()
	} else if err != nil {
		checkError(err, 1)
	}

	return string(output), exitCode, err
}

func findZSH() string {
	path, err := exec.LookPath("zsh")
	checkError(err, 1)

	return path
}

func runAllPatterns() []result {
	successfulPatterns := make([]result, 0)

	var wg sync.WaitGroup
	ch := make(chan result, 0)

	for oldPattern, newPattern := range config.Patterns {
		wg.Add(1)

		go func(oldPattern, newPattern string) {
			defer wg.Done()
			output, exitCode, _ := runZSH(fmt.Sprintf(RunAllPatternsCmd, oldPattern, newPattern))

			if exitCode != 0 {
				return
			}

			ch <- result{
				oldPattern: oldPattern,
				newPattern: newPattern,
				samples:    strings.Split(output, "\n"),
			}
		}(oldPattern, newPattern)
	}

	go func() {
		for result := range ch {
			successfulPatterns = append(successfulPatterns, result)
		}
	}()

	wg.Wait()

	return successfulPatterns
}

func choosePattern(successfulPatterns []result) result {
	for index, result := range successfulPatterns {
		fmt.Printf("%d. [%s -> %s] %s\n", index+1, result.oldPattern, result.newPattern, result.samples[0])
	}

	v := input("choose option: ")

	if v == "" || strings.ToLower(v) == "q" {
		os.Exit(0)
	}

	index, err := strconv.Atoi(v)
	checkError(err, 1)

	if l := len(successfulPatterns); index > l {
		fmt.Fprintf(os.Stdout, "choose between 1 and %d\n", l)
		os.Exit(1)
	}

	return successfulPatterns[index-1]
}

func runOnePattern(r result) {
	for _, sample := range r.samples {
		fmt.Println(sample)
	}

	proceed := input("proceed? y/n ")

	if strings.ToLower(proceed) != "y" {
		return
	}

	output, exitCode, err := runZSH(fmt.Sprintf(RunOnePatternsCmd, r.oldPattern, r.newPattern))
	checkError(err, exitCode)

	fmt.Println(output)
}

func saveConfig() {
	f, err := os.Create(configFile)
	checkError(err, 1)

	err = json.NewEncoder(f).Encode(config)
	checkError(err, 1)
}

func checkError(err error, exitCode int) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(exitCode)
	}
}

func input(label string) string {
	fmt.Print(label)
	reader := bufio.NewReader(os.Stdin)
	v, err := reader.ReadString('\n')
	checkError(err, 1)

	return strings.Trim(v, "\n")
}
