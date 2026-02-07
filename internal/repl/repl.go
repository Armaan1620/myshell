package repl

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"myshell/internal/builtins"
	"myshell/internal/executor"
	"myshell/internal/parser"

	"golang.org/x/sys/unix"
)

func Run() {
	// Set up signal handler for SIGINT (Ctrl+C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)

	go func() {
		for {
			<-sigChan
			// Check if there's a foreground process running
			if executor.GetCurrentFgPgid() > 0 {
				// Forward SIGINT to the foreground process group
				executor.SendSignalToFg(unix.SIGINT)
			} else {
				// No command running, just print a new prompt
				fmt.Print("\nmyshell> ")
			}
		}
	}()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("myshell> ")

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println()
			return
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		if strings.Contains(input, "|") {
			commands, background := parser.ParsePipeLineWithBackground(input)
			executor.ExecutePipelineWithBackground(commands, background, input)
			continue
		}

		tokens, background := parser.ParseWithBackground(input)

		if builtins.Handle(tokens) {
			continue
		}

		executor.ExecuteWithBackground(tokens, background, input)
	}

}
