package rego

import "testing"

const (
	validRegoRule = `package upswake
default wake := false
wake if {
    true
}`
	invalidRegoRule = `package upswake
default wake := false
wake if {
	RETURN TRUE
}`
	invalidPackageNameRule = `package wrongname
default wake := false
wake if {
    true
}`
	validJSON   = `{"foo": "bar"}`
	invalidJSON = `{"foo": "bar" "baz"}`
)

func TestEvaluateExpression(t *testing.T) {
	type args struct {
		rawJSON  string
		regoRule string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Valid Rego Rule and Valid JSON",
			args: args{
				rawJSON:  validJSON,
				regoRule: validRegoRule,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Valid Rego Rule and Invalid JSON",
			args: args{
				rawJSON:  invalidJSON,
				regoRule: validRegoRule,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "Invalid Rego Rule and Valid JSON",
			args: args{
				rawJSON:  validJSON,
				regoRule: invalidRegoRule,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "Invalid Rego Rule and Invalid JSON",
			args: args{
				rawJSON:  invalidJSON,
				regoRule: invalidRegoRule,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "Invalid Package Name Rego Rule and Valid JSON",
			args: args{
				rawJSON:  validJSON,
				regoRule: invalidPackageNameRule,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "UPS server data positive",
			args: args{
				rawJSON: `[{"Name":"cyberpower900","Description":"Unavailable","Master":false,"NumberOfLogins":0,"Clients":[],"Variables":[{"Name":"battery.charge","Value":100,"Type":"INTEGER","Description":"Battery charge (percent of full)","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"},{"Name":"ups.status","Value":"OL","Type":"STRING","Description":"UPS status","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"}]}]`,
				regoRule: `package upswake
default wake := false
wake if {
	input[i].Name == "cyberpower900"
	input[i].Variables[j].Name == "battery.charge"
	input[i].Variables[j].Value == 100
}`,
			},
			wantErr: false,
			want:    true,
		},
		{
			name: "UPS server data negative",
			args: args{
				rawJSON: `[{"Name":"cyberpower900","Description":"Unavailable","Master":false,"NumberOfLogins":0,"Clients":[],"Variables":[{"Name":"battery.charge","Value":50,"Type":"INTEGER","Description":"Battery charge (percent of full)","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"},{"Name":"ups.status","Value":"OL","Type":"STRING","Description":"UPS status","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"}]}]`,
				regoRule: `package upswake
default wake := false
wake if {
	input[i].Name == "cyberpower900"
	input[i].Variables[j].Name == "battery.charge"
	input[i].Variables[j].Value == 100
}`,
			},
			wantErr: false,
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EvaluateExpression(tt.args.rawJSON, tt.args.regoRule)
			if (err != nil) != tt.wantErr {
				t.Errorf("EvaluateExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("EvaluateExpression() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidRego(t *testing.T) {
	type args struct {
		rego string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Valid Rego Rule",
			args: args{
				rego: validRegoRule,
			},
			wantErr: false,
		},
		{
			name: "Invalid Rego Rule",
			args: args{
				rego: invalidRegoRule,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := IsValidRego(tt.args.rego); (err != nil) != tt.wantErr {
				t.Errorf("IsValidRego() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
