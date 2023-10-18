package util

import (
	"fmt"
	"io"
	"io/fs"
)

func GetFile(fileSystem fs.FS, fileName string) ([]byte, error) {
	file, err := fileSystem.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("could not open file %s: %w", fileName, err)
	}

	defer file.Close()
	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s: %w", fileName, err)
	}
	return fileData, nil
}
