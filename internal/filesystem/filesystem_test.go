package filesystem

import (
	"reflect"
	"testing"

	"github.com/spf13/afero"
)

func newMemFS(t *testing.T, data map[string][]byte) afero.Fs {
	t.Helper()
	memfs := afero.NewMemMapFs()

	for x := range data {
		err := afero.WriteFile(memfs, x, data[x], 0o644)
		if err != nil {
			t.Fatalf("could not write file to memfs: %s", err)
		}
	}
	return memfs
}

func TestGetFile(t *testing.T) {
	type args struct {
		fileSystem afero.Fs
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
