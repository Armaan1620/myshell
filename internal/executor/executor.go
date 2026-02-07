package executor

import (
	"fmt"
	"os"
	"os/exec"
	"sync"

	"golang.org/x/sys/unix"
)

type Job struct {
	ID    int
	PGID  int
	Cmd   string
	State string // "Running" or "Done"
}

var (
	currentFgPgid int = -1
	fgMutex       sync.RWMutex

	jobs      []Job
	jobsMutex sync.RWMutex
	nextJobID int = 1
)

func Execute(tokens []string) {
	ExecuteWithBackground(tokens, false, "")
}

func ExecuteWithBackground(tokens []string, background bool, cmdStr string) {
	if len(tokens) == 0 {
		return
	}

	clean, inFile, outFile, appendMode := parseRedirection(tokens)
	if len(clean) == 0 {
		return
	}

	cmd := exec.Command(clean[0], clean[1:]...)

	cmd.SysProcAttr = &unix.SysProcAttr{
		Setpgid: true,
	}

	// stdin
	if inFile != "" {
		f, err := os.Open(inFile)
		if err != nil {
			fmt.Println("input redirect error:", err)
			return
		}
		defer f.Close()
		cmd.Stdin = f
	} else if background {
		// Background jobs should not read from terminal
		devNull, err := os.Open("/dev/null")
		if err != nil {
			fmt.Println("error opening /dev/null:", err)
			return
		}
		defer devNull.Close()
		cmd.Stdin = devNull
	} else {
		cmd.Stdin = os.Stdin
	}

	// stdout
	if outFile != "" {
		var f *os.File
		var err error

		if appendMode {
			f, err = os.OpenFile(outFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		} else {
			f, err = os.Create(outFile)
		}

		if err != nil {
			fmt.Println("output redirect error:", err)
			return
		}
		defer f.Close()
		cmd.Stdout = f
	} else {
		cmd.Stdout = os.Stdout
	}

	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Println("start error:", err)
		return
	}

	if background {
		// Background job: print PID, track job, and return immediately
		pgid, _ := unix.Getpgid(cmd.Process.Pid)
		fmt.Printf("[bg] pid=%d\n", cmd.Process.Pid)
		addJob(pgid, cmdStr)
		// Monitor job completion in background
		go monitorJob(cmd, pgid)
		return
	}

	// Foreground job: existing behavior
	pgid, _ := unix.Getpgid(cmd.Process.Pid)
	setForeground(pgid)
	setCurrentFgPgid(pgid)

	err := cmd.Wait()

	// give terminal back to shell
	setForeground(unix.Getpgrp())
	setCurrentFgPgid(-1)

	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			fmt.Println("command error:", err)
		}
	}

}

func ExecutePipeline(commands [][]string) {
	ExecutePipelineWithBackground(commands, false, "")
}

func ExecutePipelineWithBackground(commands [][]string, background bool, cmdStr string) {
	if len(commands) == 0 {
		return
	}

	var cmds []*exec.Cmd
	var prevReader *os.File

	for i, tokens := range commands {
		if len(tokens) == 0 {
			continue
		}

		clean := tokens
		var inFile, outFile string
		var appendMode bool

		if i == 0 || i == len(commands)-1 {
			clean, inFile, outFile, appendMode = parseRedirection(tokens)
			if len(clean) == 0 {
				return
			}
		}

		cmd := exec.Command(clean[0], clean[1:]...)

		cmd.SysProcAttr = &unix.SysProcAttr{
			Setpgid: true,
		}

		// ----- stdin -----
		if i == 0 {
			if inFile != "" {
				f, err := os.Open(inFile)
				if err != nil {
					fmt.Println("input redirect error:", err)
					return
				}
				cmd.Stdin = f
				defer f.Close()
			} else if background {
				// Background pipelines should not read from terminal
				devNull, err := os.Open("/dev/null")
				if err != nil {
					fmt.Println("error opening /dev/null:", err)
					return
				}
				cmd.Stdin = devNull
				defer devNull.Close()
			} else {
				cmd.Stdin = os.Stdin
			}
		} else {
			cmd.Stdin = prevReader
		}

		var pipeReader *os.File
		var pipeWriter *os.File

		// ----- stdout -----
		if i == len(commands)-1 {
			if outFile != "" {
				var f *os.File
				var err error
				if appendMode {
					f, err = os.OpenFile(outFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				} else {
					f, err = os.Create(outFile)
				}
				if err != nil {
					fmt.Println("output redirect error:", err)
					return
				}
				cmd.Stdout = f
				defer f.Close()
			} else {
				cmd.Stdout = os.Stdout
			}
		} else {
			var err error
			pipeReader, pipeWriter, err = os.Pipe()
			if err != nil {
				fmt.Println("pipe error:", err)
				return
			}
			cmd.Stdout = pipeWriter
		}

		cmd.Stderr = os.Stderr

		if err := cmd.Start(); err != nil {
			fmt.Println("start error:", err)
			return
		}

		if pipeWriter != nil {
			pipeWriter.Close()
		}
		if prevReader != nil {
			prevReader.Close()
		}

		prevReader = pipeReader
		cmds = append(cmds, cmd)
	}

	if background {
		// Background pipeline: print PID, track job, and return immediately
		pgid, _ := unix.Getpgid(cmds[0].Process.Pid)
		fmt.Printf("[bg] pid=%d\n", cmds[0].Process.Pid)
		addJob(pgid, cmdStr)
		// Monitor job completion in background
		go monitorPipeline(cmds, pgid)
		return
	}

	// ✅ Give terminal to the pipeline (foreground job)
	pgid, _ := unix.Getpgid(cmds[0].Process.Pid)
	setForeground(pgid)
	setCurrentFgPgid(pgid)

	// ✅ Wait exactly once
	for _, cmd := range cmds {
		cmd.Wait()
	}

	// ✅ Give terminal back to shell
	setForeground(unix.Getpgrp())
	setCurrentFgPgid(-1)
}

func parseRedirection(tokens []string) (clean []string, inFile string, outFile string, appendMode bool) {
	for i := 0; i < len(tokens); i++ {
		switch tokens[i] {
		case ">":
			if i+1 < len(tokens) {
				outFile = tokens[i+1]
				appendMode = false
				i++
			}
		case ">>":
			if i+1 < len(tokens) {
				outFile = tokens[i+1]
				appendMode = true
				i++
			}
		case "<":
			if i+1 < len(tokens) {
				inFile = tokens[i+1]
				i++
			}
		default:
			clean = append(clean, tokens[i])
		}
	}
	return
}

func setForeground(pgid int) {
	fd := int(os.Stdin.Fd())
	_ = unix.IoctlSetInt(fd, unix.TIOCSPGRP, pgid)
}

func setCurrentFgPgid(pgid int) {
	fgMutex.Lock()
	defer fgMutex.Unlock()
	currentFgPgid = pgid
}

func GetCurrentFgPgid() int {
	fgMutex.RLock()
	defer fgMutex.RUnlock()
	return currentFgPgid
}

func SendSignalToFg(sig unix.Signal) {
	fgMutex.RLock()
	pgid := currentFgPgid
	fgMutex.RUnlock()

	if pgid > 0 {
		_ = unix.Kill(-pgid, sig)
	}
}

func addJob(pgid int, cmd string) {
	jobsMutex.Lock()
	defer jobsMutex.Unlock()

	job := Job{
		ID:    nextJobID,
		PGID:  pgid,
		Cmd:   cmd,
		State: "Running",
	}
	jobs = append(jobs, job)
	nextJobID++
}

func markJobDone(pgid int) {
	jobsMutex.Lock()
	defer jobsMutex.Unlock()

	for i := range jobs {
		if jobs[i].PGID == pgid {
			jobs[i].State = "Done"
			break
		}
	}
}

func GetJobs() []Job {
	jobsMutex.RLock()
	defer jobsMutex.RUnlock()

	// Return a copy of the jobs slice
	result := make([]Job, len(jobs))
	copy(result, jobs)
	return result
}

func monitorJob(cmd *exec.Cmd, pgid int) {
	// Wait for the process to finish
	cmd.Wait()
	markJobDone(pgid)
}

func monitorPipeline(cmds []*exec.Cmd, pgid int) {
	// Wait for all commands in the pipeline to finish
	for _, cmd := range cmds {
		cmd.Wait()
	}
	markJobDone(pgid)
}
