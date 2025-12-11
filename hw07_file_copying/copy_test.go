package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCopy(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := "testdata/input.txt"
	destFile := filepath.Join(tmpDir, "copy")

	tests := []struct {
		name          string
		fromPath      string
		toPath        string
		offset        int64
		limit         int64
		expectedError error
		expectedFile  string
	}{
		{
			name:          "Copy offset 0 limit 0",
			fromPath:      testFile,
			toPath:        destFile + "_offset0_limit0.txt",
			offset:        0,
			limit:         0,
			expectedError: nil,
			expectedFile:  "testdata/out_offset0_limit0.txt",
		},
		{
			name:          "Copy offset 0 limit 10",
			fromPath:      testFile,
			toPath:        destFile + "_offset0_limit10.txt",
			offset:        0,
			limit:         10,
			expectedError: nil,
			expectedFile:  "testdata/out_offset0_limit10.txt",
		},
		{
			name:          "Copy offset 0 limit 1000",
			fromPath:      testFile,
			toPath:        destFile + "_offset0_limit1000.txt",
			offset:        0,
			limit:         1000,
			expectedError: nil,
			expectedFile:  "testdata/out_offset0_limit1000.txt",
		},
		{
			name:          "Copy offset 0 limit 10000",
			fromPath:      testFile,
			toPath:        destFile + "_offset0_limit10000.txt",
			offset:        0,
			limit:         10000,
			expectedError: nil,
			expectedFile:  "testdata/out_offset0_limit10000.txt",
		},
		{
			name:          "Copy offset 100 limit 1000",
			fromPath:      testFile,
			toPath:        destFile + "_offset100_limit1000.txt",
			offset:        100,
			limit:         1000,
			expectedError: nil,
			expectedFile:  "testdata/out_offset100_limit1000.txt",
		},
		{
			name:          "Copy offset 6000 limit 1000",
			fromPath:      testFile,
			toPath:        destFile + "_offset6000_limit1000.txt",
			offset:        6000,
			limit:         1000,
			expectedError: nil,
			expectedFile:  "testdata/out_offset6000_limit1000.txt",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := Copy(tc.fromPath, tc.toPath, tc.offset, tc.limit)
			if tc.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tc.expectedError)
				} else if !errors.Is(err, tc.expectedError) && err.Error() != tc.expectedError.Error() {
					t.Errorf("expected error %v, got %v", tc.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			content, err := os.ReadFile(tc.toPath)
			if err != nil {
				t.Errorf("failed to read destination file: %v", err)
				return
			}

			expected, err := os.ReadFile(tc.expectedFile)
			if err != nil {
				t.Errorf("failed to read destination file: %v", err)
				return
			}

			if string(content) != string(expected) {
				t.Errorf("content mismatch\nexpected: %q\ngot: %q",
					string(expected), string(content))
			}

			info, err := os.Stat(tc.toPath)
			if err != nil {
				t.Errorf("failed to stat destination file: %v", err)
				return
			}

			expectedInfo, err := os.Stat(tc.expectedFile)
			if err != nil {
				t.Errorf("failed to stat destination file: %v", err)
				return
			}

			if info.Size() != expectedInfo.Size() {
				t.Errorf("file size mismatch: expected %d, got %d",
					expectedInfo.Size(), info.Size())
			}

			os.Remove(tc.toPath)
		})
	}
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
