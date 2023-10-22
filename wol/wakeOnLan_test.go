package wol

import (
	"io"
	"reflect"
	"testing"
)

type readWriteCloser struct {
	BS []byte
}

func (rwc *readWriteCloser) Read(p []byte) (n int, err error) {
	copy(p, rwc.BS)
	return len(rwc.BS), nil
}

func (rwc *readWriteCloser) Write(p []byte) (n int, err error) {
	rwc.BS = p
	return len(p), nil
}

func (rwc *readWriteCloser) Close() error {
	return nil
}

func newReadWriteCloser() io.ReadWriteCloser {
	return &readWriteCloser{
		BS: []byte{},
	}
}

func Test_newMagicPacket(t *testing.T) {
	type args struct {
		mac string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "invalid MAC",
			args: args{
				mac: "invalid",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid MAC too long",
			args: args{
				mac: "01:02:03:04:05:06:07",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid MAC too short",
			args: args{
				mac: "01:02:03:04:05",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid MAC wrong format",
			args: args{
				mac: "01:02:03:04:gg",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "valid MAC",
			args: args{
				mac: "01:02:03:04:05:06",
			},
			want: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newMagicPacket(tt.args.mac)
			if (err != nil) != tt.wantErr {
				t.Errorf("newMagicPacket() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newMagicPacket() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_wakeInternal(t *testing.T) {
	type args struct {
		dst io.ReadWriteCloser
		mac string
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		wantSent []byte
	}{
		{
			name: "invalid MAC",
			args: args{
				dst: newReadWriteCloser(),
				mac: "invalid",
			},
			wantErr:  true,
			wantSent: nil,
		},
		{
			name: "invalid MAC too long",
			args: args{
				dst: newReadWriteCloser(),
				mac: "01:02:03:04:05:06:07",
			},
			wantErr:  true,
			wantSent: nil,
		},
		{
			name: "invalid MAC too short",
			args: args{
				dst: newReadWriteCloser(),
				mac: "01:02:03:04:05",
			},
			wantErr:  true,
			wantSent: nil,
		},
		{
			name: "invalid MAC wrong format",
			args: args{
				dst: newReadWriteCloser(),
				mac: "01:02:03:04:gg",
			},
			wantErr:  true,
			wantSent: nil,
		},
		{
			name: "valid MAC",
			args: args{
				dst: newReadWriteCloser(),
				mac: "01:02:03:04:05:06",
			},
			wantErr: false,
			wantSent: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06,
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06,
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06,
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06,
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06,
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06,
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06,
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06,
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06,
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06,
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06,
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06,
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06,
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06,
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06,
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := wakeInternal(tt.args.dst, tt.args.mac); (err != nil) != tt.wantErr {
				t.Errorf("wakeInternal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			var sent = make([]byte, MagicPacketSize)
			_, err := tt.args.dst.Read(sent)
			if err != nil {
				t.Errorf("wakeInternal() error reading from dst = %v", err)
			}
			if !reflect.DeepEqual(sent, tt.wantSent) {
				t.Errorf("wakeInternal() got = %v, want %v", sent, tt.wantSent)
			}
		})
	}
}
