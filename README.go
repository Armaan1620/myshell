🐚 MyShell — A Unix-like Shell in Go

MyShell is a Unix-like command-line shell written from scratch in Go, built to deeply understand how shells actually work under the hood.

This is not a toy shell — it focuses on correct process execution, pipes, redirection, signal handling, and job control using a Go-appropriate systems design.

✨ Features
Core Shell

Interactive REPL

Built-in commands:

cd

pwd

exit

External command execution using os/exec

Pipelines

Full support for Unix pipelines (|)

Correct file descriptor handling

No deadlocks or broken pipes

ls | grep go | wc -l

Redirection

Input redirection: <

Output overwrite: >

Output append: >>

Works with pipelines

cat < file.txt | grep error > out.txt

Signal Handling (Ctrl+C)

Robust Ctrl+C handling using explicit signal forwarding

Foreground jobs receive SIGINT

Shell process remains alive

Background jobs are unaffected

Background Jobs

Background execution using &

Shell does not block on background jobs

Foreground and background jobs coexist correctly

sleep 10 &

Job Control (In Progress)

Job tracking infrastructure

jobs builtin (in progress)

Planned support for fg, bg, and Ctrl+Z

🧠 Design Philosophy

This project prioritizes correctness and clarity over shortcuts.

Key architectural decisions:

Clear separation of concerns:

repl → user interaction & prompt

parser → syntax (|, &)

executor → processes, pipes, redirection

builtins → shell-internal commands

Signal forwarding model instead of POSIX-style signal inheritance

Required for correctness in Go

Avoids fragile signal.Ignore/Reset patterns

Explicit foreground process group tracking

This design mirrors how real-world Go tools (e.g., Docker, kubectl) manage processes and signals.

📂 Project Structure
myshell/
├── main.go
├── internal/
│   ├── repl/        # REPL loop & prompt
│   ├── parser/      # Command parsing (&, |)
│   ├── executor/    # Process execution, pipes, redirection
│   ├── builtins/    # cd, pwd, exit

▶️ Running the Shell

⚠️ Do not use go run.
Interactive shells must be executed directly.

go build -o myshell
./myshell

🧪 Example Commands
# Foreground execution
sleep 5

# Ctrl+C kills foreground job
sleep 10

# Pipelines
ps aux | grep root | wc -l

# Redirection
ls > out.txt
cat < out.txt

# Background jobs
sleep 10 &

🚧 Roadmap

 REPL & command execution

 Pipes and redirection

 Ctrl+C handling

 Background jobs (&)

 jobs builtin

 fg / bg

 Ctrl+Z (SIGTSTP)

 Command history

 Quoting & escaping

🎓 What I Learned

How Unix shells actually manage processes

Why built-in commands must run in the shell process

How pipelines work at the file descriptor level

How signals are delivered and handled

Why Go requires a different approach to job control than C

The difference between “working” and “correct”

📌 Motivation

This project was built as a systems engineering learning exercise to better understand operating systems, process management, and low-level behavior that underpins backend and infrastructure software.
