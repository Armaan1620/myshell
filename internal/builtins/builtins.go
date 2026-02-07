package builtins

import (
	"fmt"
	"os"

	"myshell/internal/executor"
)

func Handle(tokens []string) bool {
	if len(tokens) == 0 {
		return true
	}

	switch tokens[0] {
	case "cd":
		return cd(tokens)
	case "pwd":
		return pwd()
	case "jobs":
		return jobs()
	case "exit":
		os.Exit(0)
	default:
		return false
	}
	return false
}

func cd(tokens []string) bool {
	if len(tokens) < 2 {
		fmt.Println("cd: missing argument")
		return true
	}

	err := os.Chdir(tokens[1])
	if err != nil {
		fmt.Println("cd: ", err)
	}
	return true
}

func pwd() bool {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("pwd: ", err)
		return true
	}
	fmt.Println(dir)
	return true
}

func jobs() bool {
	jobList := executor.GetJobs()
	for _, job := range jobList {
		fmt.Printf("[%d] %-7s %s\n", job.ID, job.State, job.Cmd)
	}
	return true
}
