package utils

import (
	"io"
	"mime/multipart"
	"os"
)

func SaveUploadedFile(fileData *multipart.FileHeader, dst string) error {
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
