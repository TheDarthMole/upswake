package util

import (
	"github.com/spf13/afero"
)

func GetFile(fileSystem afero.Fs, fileName string) ([]byte, error) {
	return afero.ReadFile(fileSystem, fileName)
}
