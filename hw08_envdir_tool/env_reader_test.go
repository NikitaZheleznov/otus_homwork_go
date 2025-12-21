package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadDir(t *testing.T) {
	// Путь к тестовым данным
	testDataDir := filepath.Join("testdata", "env")

	// Проверяем существование директории
	if _, err := os.Stat(testDataDir); os.IsNotExist(err) {
		t.Skipf("Test data directory not found: %s", testDataDir)
	}

	// Запускаем ReadDir
	env, err := ReadDir(testDataDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	// Проверяем результаты на основе ожидаемого содержимого файлов
	expected := Environment{
		"BAR": EnvValue{
			Value:      "bar",
			NeedRemove: false,
		},
		"EMPTY": EnvValue{
			Value:      "",
			NeedRemove: false,
		},
		"FOO": EnvValue{
			Value:      "   foo\nwith new line",
			NeedRemove: false,
		},
		"HELLO": EnvValue{
			Value:      "\"hello\"",
			NeedRemove: false,
		},
		"UNSET": EnvValue{
			Value:      "",
			NeedRemove: true,
		},
	}

	if len(env) != len(expected) {
		t.Errorf("Expected %d variables, got %d", len(expected), len(env))
	}

	for key, expectedValue := range expected {
		actualValue, ok := env[key]
		if !ok {
			t.Errorf("Missing variable: %s", key)
			continue
		}

		if actualValue.Value != expectedValue.Value {
			t.Errorf("Variable %s: expected value %q, got %q",
				key, expectedValue.Value, actualValue.Value)
		}

		if actualValue.NeedRemove != expectedValue.NeedRemove {
			t.Errorf("Variable %s: expected NeedRemove %v, got %v",
				key, expectedValue.NeedRemove, actualValue.NeedRemove)
		}
	}
}
