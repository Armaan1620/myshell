package parser

type Command struct {
	Name string
	Args []string
	Stdin string
	Stdout string
	Append bool
}