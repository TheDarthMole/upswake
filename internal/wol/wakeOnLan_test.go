package wol

import (
	"io"
	"testing"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validMagicPacketMAC = "01:02:03:04:05:06"

var validMagicPacket = []byte{
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
}

type readWriteCloser struct {
	BS []byte
}

type readWriteCloserError struct {
	readWriteCloser
}

func (rwc *readWriteCloser) Read(p []byte) (n int, err error) {
	copy(p, rwc.BS)
	return len(rwc.BS), nil
}

func (rwc *readWriteCloser) Write(p []byte) (n int, err error) {
	rwc.BS = p
	return len(p), nil
}

func (*readWriteCloser) Close() error {
	return nil
}

func (*readWriteCloserError) Write(_ []byte) (int, error) {
	return 15, nil
}

func newValidTestWoLTarget() *entity.TargetServer {
	return &entity.TargetServer{
		Name:      "test",
		MAC:       "01:02:03:04:05:06",
		Broadcast: "127.0.0.255",
		Port:      9,
		Interval:  "15m",
		Rules:     []string{},
	}
}

func newReadWriteCloser() io.ReadWriteCloser {
	return &readWriteCloser{
		BS: []byte{},
	}
}

func newReadWriteCloserError() io.ReadWriteCloser {
	return &readWriteCloserError{
		readWriteCloser{
			BS: []byte{},
		},
	}
}

func Test_newMagicPacket(t *testing.T) {
	type args struct {
		mac string
	}
	tests := []struct {
		name  string
		args  args
		want  []byte
		error error
	}{
		{
			name: "invalid MAC",
			args: args{
				mac: "invalid",
			},
			want:  nil,
			error: ErrFailedCreateMagicPacket,
		},
		{
			name: "invalid MAC too long",
			args: args{
				mac: "01:02:03:04:05:06:07",
			},
			want:  nil,
			error: ErrFailedCreateMagicPacket,
		},
		{
			name: "invalid MAC too short",
			args: args{
				mac: "01:02:03:04:05",
			},
			want:  nil,
			error: ErrFailedCreateMagicPacket,
		},
		{
			name: "invalid MAC wrong format",
			args: args{
				mac: "01:02:03:04:gg",
			},
			want:  nil,
			error: ErrFailedCreateMagicPacket,
		},
		{
			name: "valid MAC",
			args: args{
				mac: validMagicPacketMAC,
			},
			want:  validMagicPacket,
			error: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newMagicPacket(tt.args.mac)

			assert.ErrorIs(t, err, tt.error)
			assert.Equal(t, tt.want, got)
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
		error    error
		wantSent []byte
	}{
		{
			name: "invalid MAC",
			args: args{
				dst: newReadWriteCloser(),
				mac: "invalid",
			},
			error:    ErrFailedCreateMagicPacket,
			wantSent: nil,
		},
		{
			name: "invalid MAC too long",
			args: args{
				dst: newReadWriteCloser(),
				mac: "01:02:03:04:05:06:07",
			},
			error:    ErrFailedCreateMagicPacket,
			wantSent: nil,
		},
		{
			name: "invalid MAC too short",
			args: args{
				dst: newReadWriteCloser(),
				mac: "01:02:03:04:05",
			},
			error:    ErrFailedCreateMagicPacket,
			wantSent: nil,
		},
		{
			name: "invalid MAC wrong format",
			args: args{
				dst: newReadWriteCloser(),
				mac: "01:02:03:04:gg",
			},
			error:    ErrFailedCreateMagicPacket,
			wantSent: nil,
		},
		{
			name: "valid MAC",
			args: args{
				dst: newReadWriteCloser(),
				mac: validMagicPacketMAC,
			},
			error:    nil,
			wantSent: validMagicPacket,
		},
		{
			name: "invalid write length",
			args: args{
				dst: newReadWriteCloserError(),
				mac: validMagicPacketMAC,
			},
			error:    ErrExpectedPacketSize,
			wantSent: validMagicPacket,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := wakeInternal(tt.args.dst, tt.args.mac)
			assert.ErrorIs(t, err, tt.error)

			if tt.error != nil {
				return
			}

			sent := make([]byte, MagicPacketSize)
			_, err = tt.args.dst.Read(sent)
			require.NoError(t, err)

			assert.Equal(t, tt.wantSent, sent)
		})
	}
}

func TestNewWoLClient(t *testing.T) {
	type args struct {
		target *entity.TargetServer
	}
	tests := []struct {
		name string
		args args
		want *WakeOnLan
	}{
		{
			name: "valid target",
			args: args{
				newValidTestWoLTarget(),
			},
			want: &WakeOnLan{newValidTestWoLTarget()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewWoLClient(tt.args.target)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWakeOnLan_Wake(t *testing.T) {
	type fields struct {
		WoLTarget *entity.TargetServer
	}
	tests := []struct {
		name   string
		fields fields
		error  error
	}{
		{
			name: "valid",
			fields: fields{
				newValidTestWoLTarget(),
			},
			error: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tgt := &WakeOnLan{
				tt.fields.WoLTarget,
			}
			err := tgt.Wake()
			assert.ErrorIs(t, err, tt.error)
		})
	}
}
