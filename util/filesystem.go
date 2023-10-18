package util

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"strings"
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

func removeLastElement(input string, separator string) string {
	lastIndex := strings.LastIndex(input, separator)
	if lastIndex >= 0 {
		return input[:lastIndex]
	}
	return ""
}

func GetCurrentDirectory() string {
	cwd, err := os.Executable()
	if err != nil {
		log.Fatalf("could not get current working directory: %s", err)
	}
	return removeLastElement(cwd, "/")
}

func CreateFile(file string, data []byte) error {
	f, err := os.Create(file)
	if err != nil {
		log.Fatalf("config file could not be created: %s", err)
	}
	i, err := f.Write(data)
	if err != nil {
		return err
	}
	if i != len(data) {
		return fmt.Errorf("could not write all data to file")
	}
	return nil
}
