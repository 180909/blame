package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

func main() {
	var ignore string = "default"

	args := os.Args[1:]
	if len(args) != 0 {
		for _, arg := range args {
			if strings.HasPrefix(arg, "-g") {
				ignores := strings.Split(args[0], "=")
				ignore = ignores[len(ignores)-1]
			}
		}
	}

	o := os.Stdout
	defer func() {
		o.Close()
		os.Exit(0)
	}()

	authors, fileNames := make(map[string]int), new([]string)
	wd, _ := os.Getwd()

	getAllFileName(wd, fileNames)
	wd = wd + "/"

	for _, name := range *fileNames {
		fileName := strings.TrimPrefix(name, wd)

		if strings.HasPrefix(fileName, ".") {
			continue
		}

		if ignore != "default" {
			if !strings.HasPrefix(fileName, ignore) {
				count(gitBlame(fileName), authors)
			}
		} else {
			count(gitBlame(fileName), authors)
		}
	}

	for author, times := range authors {
		o.WriteString(fmt.Sprintf("author: %s, lines: %d \n", author, times))
	}
}

func gitBlame(fileName string) []string {
	args := []string{"blame", "--root", fileName}
	cmd := exec.Command("git", args...)

	cmdLineArgs := strings.Join(args, " ")
	fmt.Printf("git %s\n", cmdLineArgs)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil
	}
	return strings.Split(string(out), "\n")
}

func count(lines []string, authors map[string]int) {
	for _, line := range lines {
		if line == "" {
			continue
		}
		tmps := strings.Split(line, "(")
		tmps = strings.Split(tmps[1], ")")
		tmps = strings.Split(tmps[0], " ")
		authors[tmps[0]]++
	}
	return
}

func getAllFileName(prefix string, fileNames *[]string) {
	fs, err := ioutil.ReadDir(prefix)
	if err != nil {
		os.Stderr.WriteString("blame: " + err.Error() + "\n")
		os.Exit(1)
	}

	for _, f := range fs {
		if f.IsDir() {
			getAllFileName(prefix+"/"+f.Name(), fileNames)
		} else {
			*fileNames = append(*fileNames, prefix+"/"+f.Name())
		}
	}
	return
}

// check arg exist or is file or is dir
// if arg not exist, exit program
// if arg is dir, return 1
// if arg is file, return 2
func checkFileType(fileName string, o io.StringWriter) int {
	s, err := os.Stat(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			o.WriteString("\nblame: no such file or directory: " + fileName + " \n")

			os.Exit(0)
		}
		e := os.Stderr
		e.WriteString(err.Error())
		e.Close()
		os.Exit(1)
	}
	if s.IsDir() {
		return 1
	} else {
		return 0
	}
}
