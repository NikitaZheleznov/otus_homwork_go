package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	env := make(Environment)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return env, fmt.Errorf("directory does not exist: %s", dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return env, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			continue
		}

		filename := entry.Name()

		if strings.Contains(filename, "=") {
			return env, fmt.Errorf("invalid filename contains '=': %s", filename)
		}

		envValue, err := readEnvFile(filepath.Join(dir, filename))
		if err != nil {
			return env, fmt.Errorf("failed to read file %s: %w", filename, err)
		}

		env[filename] = envValue
	}

	return env, nil
}

func readEnvFile(filepath string) (EnvValue, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return EnvValue{}, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return EnvValue{}, err
	}

	if fileInfo.Size() == 0 {
		return EnvValue{NeedRemove: true}, nil
	}

	reader := bufio.NewReader(file)

	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return EnvValue{}, err
	}

	line = strings.TrimSuffix(line, "\n")

	line = strings.TrimRight(line, " \t")

	line = string(bytes.ReplaceAll([]byte(line), []byte{0x00}, []byte("\n")))

	return EnvValue{
		Value:      line,
		NeedRemove: false,
	}, nil
}
