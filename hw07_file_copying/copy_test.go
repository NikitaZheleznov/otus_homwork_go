package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := "testdata/input.txt"
	destFile := filepath.Join(tmpDir, "copy")

	tests := []struct {
		name         string
		fromPath     string
		toPath       string
		offset       int64
		limit        int64
		expectedFile string
	}{
		{
			name:         "Copy offset 0 limit 0",
			fromPath:     testFile,
			toPath:       destFile + "_offset0_limit0.txt",
			offset:       0,
			limit:        0,
			expectedFile: "testdata/out_offset0_limit0.txt",
		},
		{
			name:         "Copy offset 0 limit 10",
			fromPath:     testFile,
			toPath:       destFile + "_offset0_limit10.txt",
			offset:       0,
			limit:        10,
			expectedFile: "testdata/out_offset0_limit10.txt",
		},
		{
			name:         "Copy offset 0 limit 1000",
			fromPath:     testFile,
			toPath:       destFile + "_offset0_limit1000.txt",
			offset:       0,
			limit:        1000,
			expectedFile: "testdata/out_offset0_limit1000.txt",
		},
		{
			name:         "Copy offset 0 limit 10000",
			fromPath:     testFile,
			toPath:       destFile + "_offset0_limit10000.txt",
			offset:       0,
			limit:        10000,
			expectedFile: "testdata/out_offset0_limit10000.txt",
		},
		{
			name:         "Copy offset 100 limit 1000",
			fromPath:     testFile,
			toPath:       destFile + "_offset100_limit1000.txt",
			offset:       100,
			limit:        1000,
			expectedFile: "testdata/out_offset100_limit1000.txt",
		},
		{
			name:         "Copy offset 6000 limit 1000",
			fromPath:     testFile,
			toPath:       destFile + "_offset6000_limit1000.txt",
			offset:       6000,
			limit:        1000,
			expectedFile: "testdata/out_offset6000_limit1000.txt",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := Copy(tc.fromPath, tc.toPath, tc.offset, tc.limit)
			require.NoError(t, err)
			assertFilesEqual(t, tc.toPath, tc.expectedFile)
		})
	}
}

func assertFilesEqual(t *testing.T, actual, expected string) {
	t.Helper()

	actualContent, err := os.ReadFile(actual)
	require.NoError(t, err)

	expectedContent, err := os.ReadFile(expected)
	require.NoError(t, err)

	require.Equal(t, string(expectedContent), string(actualContent))

	actualInfo, err := os.Stat(actual)
	require.NoError(t, err)

	expectedInfo, err := os.Stat(expected)
	require.NoError(t, err)

	require.Equal(t, expectedInfo.Size(), actualInfo.Size())
}

func TestCopyEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()

	emptyFile := filepath.Join(tmpDir, "empty.txt")
	destFile := filepath.Join(tmpDir, "copy.txt")

	file, err := os.Create(emptyFile)
	if err != nil {
		t.Fatalf("failed to create empty file: %v", err)
	}
	file.Close()

	err = Copy(emptyFile, destFile, 0, 0)
	errors.Is(err, ErrUnsupportedFile)
}

func TestOffsetMoreThanFileSize(t *testing.T) {
	tmpDir := t.TempDir()
	input := "testdata/input.txt"
	destFile := filepath.Join(tmpDir, "copy.txt")

	errors.Is(Copy(input, destFile, 1000000, 0), ErrOffsetExceedsFileSize)
}
