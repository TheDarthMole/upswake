package wol

import "testing"

func TestWake(t *testing.T) {
	type args struct {
		mac string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Wake(tt.args.mac); (err != nil) != tt.wantErr {
				t.Errorf("Wake() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
