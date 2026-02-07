package parser

import "strings"

func Parse(input string) []string {
	return strings.Fields(input)
}

// ParseWithBackground detects trailing & and returns tokens and background flag
func ParseWithBackground(input string) ([]string, bool) {
	tokens := strings.Fields(input)
	if len(tokens) > 0 && tokens[len(tokens)-1] == "&" {
		return tokens[:len(tokens)-1], true
	}
	return tokens, false
}

func ParsePipeLine(input string) [][]string {
	parts := strings.Split(input, "|")

	var commands [][]string
	for _, part := range parts {
		part := strings.TrimSpace(part)
		if part == "" {
			continue
		}
		commands = append(commands, strings.Fields(part))
	}

	return commands
}

// ParsePipeLineWithBackground detects trailing & and returns commands and background flag
func ParsePipeLineWithBackground(input string) ([][]string, bool) {
	trimmed := strings.TrimSpace(input)
	background := strings.HasSuffix(trimmed, "&")

	if background {
		// Remove trailing & before parsing
		trimmed = strings.TrimSuffix(strings.TrimSpace(trimmed), "&")
		trimmed = strings.TrimSpace(trimmed)
	}

	parts := strings.Split(trimmed, "|")
	var commands [][]string
	for _, part := range parts {
		part := strings.TrimSpace(part)
		if part == "" {
			continue
		}
		commands = append(commands, strings.Fields(part))
	}

	return commands, background
}

func ParseCommand(input string) Command {
	tokens := strings.Fields(input)

	var cmd Command
	var args []string

	for i := 0; i < len(tokens); i++ {
		switch tokens[i] {
		case ">":
			cmd.Stdout = tokens[i+1]
			cmd.Append = false
			i++
		case ">>":
			cmd.Stdout = tokens[i+1]
			cmd.Append = true
			i++
		case "<":
			cmd.Stdin = tokens[i+1]
			i++
		default:
			args = append(args, tokens[i+1])
		}
	}
	cmd.Name = args[0]
	cmd.Args = args[1:]
	return cmd
}
