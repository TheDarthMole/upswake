package logging

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
			name: "empty string",
			args: args{
				input: "",
			},
			want: "",
		},
		{
			name: "single character",
			args: args{
				input: "a",
			},
			want: "a",
		},
		{
			name: "single quote",
			args: args{
				input: `"`,
			},
			want: "",
		},
		{
			name: "double quotes",
			args: args{
				input: `""`,
			},
			want: "",
		},
		{
			name: "string with quote",
			args: args{
				input: `this is a sentence with a "quote" in it`,
			},
			want: "this is a sentence with a quote in it",
		},
		{
			name: "newline",
			args: args{
				input: "\n",
			},
			want: "",
		},
		{
			name: "string with newline",
			args: args{
				input: "hey\n",
			},
			want: "hey",
		},
		{
			name: "multiple newlines",
			args: args{
				input: "hey\n\n",
			},
			want: "hey",
		},
		{
			name: "multiple return characters",
			args: args{
				input: "hey\r\r",
			},
			want: "hey",
		},
		{
			name: "whitespace around text",
			args: args{
				input: "   hey   ",
			},
			want: "hey",
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
