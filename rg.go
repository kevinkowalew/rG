package main

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"strings"
	"github.com/manifoldco/promptui"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("You must provide one search term for your query")
	}

	cmd := exec.Command("grep", "-r", "-n", string(os.Args[1]))
	out, err := cmd.Output()
	if string(out) == "" {
		if err != nil && err.Error() != "exit status 1" {
		   panic(err)
		}
	}

	lines := make([]string,0)
	for _,line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}

		if strings.Contains(line, "go.mod") {
			continue
		}

		if strings.Contains(line, "go.sum") {
			continue
		}

		if strings.Contains(line, ".git/") {
			continue
		}

		lines = append(lines, line)
	}

	if len(lines) == 0 {
		return 
	}

	p := promptui.Select{
		Label: "Open",
		Items: lines,
	}
	p.IsVimMode = true
	p.Size = 20
	p.HideHelp = true

	_, gr, err :=  p.Run()
	if err != nil {
	   panic(err)
   }

   pgr, err := parseGrepResult(gr)
   if err != nil {
	   panic(err)
   }

   launchVimCmd := exec.Command("vim", "+" + pgr.lineNumber, pgr.fileName)
   launchVimCmd.Stdin = os.Stdin
   launchVimCmd.Stdout = os.Stdout
   err = launchVimCmd.Run()
   if err != nil {
	   panic(err)
   }
}

type grepResult struct {
	fileName string
	lineNumber string
	target []string
}

func parseGrepResult(line string) (*grepResult, error) {
	p := strings.Split(line, ":")
	if len(p) < 3 {
		return nil, errors.New("Failed to parse grep result from line: " + line)
	}

	return &grepResult{
		fileName: p[0],
		lineNumber: p[1],
		target: p[2:],
	}, nil
}
