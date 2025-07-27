package util

import "testing"

func TestSanitizeString(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Empty string",
			args: args{input: ""},
			want: "",
		},
		{
			name: "String with spaces",
			args: args{input: "   Hello World   "},
			want: "Hello World",
		},
		{
			name: "String with newlines and carriage returns",
			args: args{input: "Hello\nWorld\r\n"},
			want: "HelloWorld",
		},
		{
			name: "String with non-printable characters",
			args: args{input: "Hello\x00World\x01"},
			want: "HelloWorld",
		},
		{
			name: "String with special characters",
			args: args{input: "Hello, World! @#$%^&*()"},
			want: "Hello, World! @#$%^&*()",
		},
		{
			name: "String with mixed content",
			args: args{input: "   Hello\x00World! \nThis is a test.   "},
			want: "HelloWorld! This is a test.",
		},
		{
			name: "String with only non-printable characters",
			args: args{input: "\x00\x01\x02\x03"},
			want: "",
		},
		{
			name: "String with leading and trailing spaces and newlines",
			args: args{input: "   \nHello World\n   "},
			want: "Hello World",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeString(tt.args.input); got != tt.want {
				t.Errorf("SanitizeString() = %v, want %v", got, tt.want)
			}
		})
	}
}
