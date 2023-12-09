package util

import (
	"fmt"
	"github.com/hack-pad/hackpadfs"
	"github.com/hack-pad/hackpadfs/mem"
	"os"
	"reflect"
	"testing"
)

func newMemFS(t *testing.T, data map[string][]byte) hackpadfs.FS {
	t.Helper()
	memfs, err := mem.NewFS()
	if err != nil {
		t.Fatalf("could not create memfs: %s", err)
	}

	for x := range data {
		err = hackpadfs.WriteFullFile(memfs, x, data[x], 0644)
		if err != nil {
			t.Fatalf("could not write file to memfs: %s", err)
		}
	}
	return memfs
}

type FaultyFS struct {
	hackpadfs.FS
	File FaultyFile
}

type FaultyCreateFS struct {
	FaultyFS
}

func (f FaultyCreateFS) Create(name string) (hackpadfs.File, error) {
	fmt.Println("Create")
	return nil, &hackpadfs.PathError{Op: "create", Path: name, Err: hackpadfs.ErrPermission}
}

type FaultyFile struct {
	hackpadfs.File
}

func (f FaultyFile) Write(p []byte) (n int, err error) {
	fmt.Println("Write")
	return 0, &hackpadfs.PathError{Op: "write", Path: "", Err: hackpadfs.ErrPermission}
}

type FaultyWriteFS struct {
	hackpadfs.FS
	File FaultyFile
}

func (f FaultyWriteFS) Create(name string) (hackpadfs.File, error) {
	fmt.Println("Create")
	return f.File, nil
}

func (f FaultyWriteFS) WriteFile(name string, data []byte, perm os.FileMode) error {
	fmt.Println("WriteFile")
	return &hackpadfs.PathError{Op: "write", Path: name, Err: hackpadfs.ErrPermission}
}

type FaultyWriteSizeFS struct {
	hackpadfs.FS
	File FaultyWriteSizeFile
}

func (f FaultyWriteSizeFS) Create(name string) (hackpadfs.File, error) {
	fmt.Println("Create")
	return f.File, nil
}

type FaultyWriteSizeFile struct {
	FaultyFile
}

func (f FaultyWriteSizeFile) Write(p []byte) (n int, err error) {
	fmt.Println("Write")
	return -1, nil
}

func TestGetFile(t *testing.T) {
	type args struct {
		fileSystem hackpadfs.FS
		fileName   string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "Valid File",
			args: args{
				fileSystem: newMemFS(t, map[string][]byte{
					"filename1.txt": []byte("data1"),
				}),
				fileName: "filename1.txt",
			},
			want:    []byte("data1"),
			wantErr: false,
		},
		{
			name: "Invalid File",
			args: args{
				fileSystem: newMemFS(t, map[string][]byte{}),
				fileName:   "doesnotexist.txt",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetFile(tt.args.fileSystem, tt.args.fileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFile() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCurrentDirectory(t *testing.T) {
	t.Run("GetCurrentDirectory", func(t *testing.T) {
		dir, err := GetCurrentDirectory()
		if err != nil {
			t.Errorf("GetCurrentDirectory() error = %v", err)
		}
		if dir == "" {
			t.Errorf("GetCurrentDirectory() dir should not be empty: %v", dir)
		}
	})
}

func TestFileExists(t *testing.T) {
	type args struct {
		fileSystem hackpadfs.FS
		file       string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Valid File",
			args: args{
				fileSystem: newMemFS(t, map[string][]byte{
					"filename1.txt": []byte("data1"),
				}),
				file: "filename1.txt",
			},
			want: true,
		},
		{
			name: "Empty File",
			args: args{
				fileSystem: newMemFS(t, map[string][]byte{
					"filename1.txt": []byte(""),
				}),
				file: "filename1.txt",
			},
			want: true,
		},
		{
			name: "Non Existent File",
			args: args{
				fileSystem: newMemFS(t, map[string][]byte{}),
				file:       "filename1.txt",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FileExists(tt.args.fileSystem, tt.args.file); got != tt.want {
				t.Errorf("FileExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetLocalFS(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Valid Local FS",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetLocalFS()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLocalFS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestCreateFile(t *testing.T) {
	type args struct {
		fsys hackpadfs.FS
		file string
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		creates int
		writes  int
	}{
		{
			name: "Valid File",
			args: args{
				fsys: newMemFS(t, map[string][]byte{}),
				file: "filename1.txt",
				data: []byte("data1"),
			},
			wantErr: false,
			writes:  1,
			creates: 1,
		},
		{
			name: "Invalid File",
			args: args{
				fsys: FaultyCreateFS{},
				file: "filename1.txt",
				data: []byte("data1"),
			},
			wantErr: true,
			creates: 1,
			writes:  0,
		},
		{
			name: "Invalid Write",
			args: args{
				fsys: FaultyWriteFS{},
				file: "filename1.txt",
				data: []byte("data1"),
			},
			wantErr: true,
			creates: 1,
			writes:  1,
		},
		{
			name: "Invalid Write Size",
			args: args{
				fsys: FaultyWriteSizeFS{},
				file: "filename1.txt",
				data: []byte("data1"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CreateFile(tt.args.fsys, tt.args.file, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			_, err = hackpadfs.Stat(tt.args.fsys, tt.args.file)
			if err != nil {
				t.Errorf("could not stat file %s: %s", tt.args.file, err)
				return
			}
			file, err := hackpadfs.ReadFile(tt.args.fsys, tt.args.file)
			if err != nil {
				t.Errorf("could not read file %s: %s", tt.args.file, err)
				return
			}
			if !reflect.DeepEqual(file, tt.args.data) {
				t.Errorf("CreateFile() got = %v, want %v", file, tt.args.data)
			}
		})
	}
}
