package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	if len(cmd) == 0 {
		fmt.Fprintf(os.Stderr, "no command specified\n")
		return 1
	}

	prepareEnvironment(env)

	command := exec.Command(cmd[0], cmd[1:]...)

	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	command.Env = os.Environ()

	if err := command.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus()
			}
		}
		fmt.Fprintf(os.Stderr, "failed to run command: %v\n", err)
		return 1
	}

	return 0
}

func prepareEnvironment(env Environment) {
	for name, envValue := range env {
		if envValue.NeedRemove {
			os.Unsetenv(name)
		} else {
			os.Setenv(name, envValue.Value)
		}
	}
}
