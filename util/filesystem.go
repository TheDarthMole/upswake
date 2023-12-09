package util

import (
	"fmt"
	"github.com/hack-pad/hackpadfs"
	hackpados "github.com/hack-pad/hackpadfs/os"
	"log"
	"os"
	"path/filepath"
)

func GetFile(fileSystem hackpadfs.FS, fileName string) ([]byte, error) {
	return hackpadfs.ReadFile(fileSystem, fileName)
}

func GetCurrentDirectory() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(ex), nil
}

func FileExists(fileSystem hackpadfs.FS, file string) bool {
	fileInfo, err := hackpadfs.Stat(fileSystem, file)

	if err != nil {
		log.Printf("could not stat file %s: %s", file, err)
		return false
	}

	return !fileInfo.IsDir()
}

func GetLocalFS() (hackpadfs.FS, error) {
	fs := hackpados.NewFS()
	cwd, err := GetCurrentDirectory()
	if err != nil {
		return nil, fmt.Errorf("could not get current directory: %s", err)
	}
	path, err := fs.FromOSPath(cwd)
	if err != nil {
		return nil, fmt.Errorf("could not convert %s to a path: %s", cwd, err)
	}
	sub, err := fs.Sub(path)
	if err != nil {
		return nil, fmt.Errorf("could not get subdirectory %s: %s", path, err)
	}
	return sub, nil
}

func CreateFile(fsys hackpadfs.FS, file string, data []byte) error {
	f, err := hackpadfs.Create(fsys, file)
	if err != nil {
		return fmt.Errorf("could not create file %s: %s", file, err)
	}
	i, err := hackpadfs.WriteFile(f, data)
	if err != nil {
		return fmt.Errorf("could not write data to file %s: %s", file, err)
	}
	if i != len(data) {
		return fmt.Errorf("could not write all data to file %s", file)
	}
	return nil
}
