package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/cheggaaa/pb/v3"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	srcFile, err := os.Open(fromPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(toPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	if offset > srcInfo.Size() {
		return ErrOffsetExceedsFileSize
	}

	if !srcInfo.Mode().IsRegular() {
		return ErrUnsupportedFile
	}

	if offset > 0 {
		_, err = srcFile.Seek(offset, io.SeekStart)
		if err != nil {
			return fmt.Errorf("failed to seek to offset: %w", err)
		}
	}

	bytesToCopy := srcInfo.Size() - offset
	if limit > 0 && limit < bytesToCopy {
		bytesToCopy = limit
	}
	err = copyWithBar(srcFile, dstFile, bytesToCopy)
	if err != nil {
		os.Remove(toPath)
		return fmt.Errorf("copy failed: %w", err)
	}

	return nil
}

func copyWithBar(src io.Reader, dst io.Writer, totalBytes int64) error {
	var err error
	bar := pb.New64(totalBytes)
	bar.Set(pb.Bytes, true)
	bar.Set("prefix", "Copy file: ")
	if totalBytes == 0 {
		return nil
	}
	reader := bar.NewProxyReader(src)
	bar.Start()
	_, err = io.CopyN(dst, reader, totalBytes)
	bar.Finish()
	return err
}
