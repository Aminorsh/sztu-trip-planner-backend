package utils

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

func SaveUploadedFile(fileData *multipart.FileHeader, dst string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	srcFile, err := fileData.Open()
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
