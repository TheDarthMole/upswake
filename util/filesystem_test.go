package util

import (
	"fmt"
	"github.com/hack-pad/hackpadfs"
	"github.com/hack-pad/hackpadfs/mem"
	"reflect"
	"testing"
)

func TestGetFile(t *testing.T) {
	filename1 := "test.txt"
	filename2 := "test2.txt"
	data1 := []byte("test")
	data2 := []byte("test2")
	memFS := newMemFS(t, map[string][]byte{
		filename1: data1,
		filename2: data2,
	})

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
				fileSystem: memFS,
				fileName:   filename1,
			},
			want:    data1,
			wantErr: false,
		},
		{
			name: "Invalid File",
			args: args{
				fileSystem: newMemFS(t, map[string][]byte{}),
				fileName:   "doesnotexist.txt",
			},
			want:    []byte(""),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetFile(tt.args.fileSystem, tt.args.fileName)
			fmt.Println(err != nil)
			fmt.Println(tt.wantErr)
			fmt.Println((err != nil) != tt.wantErr)

			if (err != nil) != tt.wantErr {
				fmt.Println("THIS WAS HIT!")
				t.Errorf("GetFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFile() got = %v, want %v", got, tt.want)
			}
		})
	}
}

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
