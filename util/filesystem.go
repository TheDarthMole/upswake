package util

import (
	"fmt"
	"io"
	"io/fs"
)

func ListFiles(fileSystem fs.FS, dir string) ([]fs.DirEntry, error) {
	readDir, err := fs.ReadDir(fileSystem, dir)
	if err != nil {
		return nil, err
	}
	return readDir, nil
}

func GetFile(fileSystem fs.FS, fileName string) ([]byte, error) {
	fmt.Println(fileName)
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
