package evaluator

import (
	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/infrastructure/config/viper"
	"github.com/hack-pad/hackpadfs"
	"github.com/hack-pad/hackpadfs/mem"
	"reflect"
	"testing"
)

const (
	validNUTOutput   = `[{"Name":"cyberpower900","Description":"Unavailable","Master":false,"NumberOfLogins":0,"Clients":[],"Variables":[{"Name":"battery.charge","Value":100,"Type":"INTEGER","Description":"Battery charge (percent of full)","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"},{"Name":"ups.status","Value":"OL","Type":"STRING","Description":"UPS status","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"}]}]`
	invalidNUTOutput = "invalid!"
)

var (
	defaultConfig, _ = viper.CreateDefaultConfig()
	tempFS, _        = mem.NewFS()
	alwaysTrueRego   = []byte("package upswake\n\ndefault wake = true")
	alwaysFalseRego  = []byte("package upswake\n\ndefault wake = false")
)

func TestNewRegoEvaluator(t *testing.T) {
	alwaysTrueRegoFS, err := mem.NewFS()
	if err != nil {
		t.Fatal("Failed to setup memfs")
	}
	if writeMemFile(alwaysTrueRegoFS, "test.rego", alwaysTrueRego, t) != nil {
		t.Fatal(err)
	}

	alwaysFalseRegoFS, err := mem.NewFS()
	if err != nil {
		t.Fatal("Failed to setup memfs")
	}
	if writeMemFile(alwaysTrueRegoFS, "test.rego", alwaysFalseRego, t) != nil {
		t.Fatal(err)
	}

	type args struct {
		config  *entity.Config
		mac     string
		rulesFS hackpadfs.FS
	}

	tests := []struct {
		name string
		args args
		want *RegoEvaluator
	}{
		{
			name: "valid config 1",
			args: args{
				config:  defaultConfig,
				mac:     "00:00:00:00:00:00",
				rulesFS: alwaysTrueRegoFS,
			},
			want: &RegoEvaluator{
				config:  defaultConfig,
				rulesFS: alwaysTrueRegoFS,
				mac:     "00:00:00:00:00:00",
			},
		},
		{
			name: "valid config 2",
			args: args{
				config:  defaultConfig,
				mac:     "00:00:00:00:00:55",
				rulesFS: alwaysFalseRegoFS,
			},
			want: &RegoEvaluator{
				config:  defaultConfig,
				rulesFS: alwaysFalseRegoFS,
				mac:     "00:00:00:00:00:55",
			},
		},
		//	TODO: Add more test cases
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRegoEvaluator(tt.args.config, tt.args.mac, tt.args.rulesFS); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRegoEvaluator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegoEvaluator_EvaluateExpressions(t *testing.T) {
	type fields struct {
		config  *entity.Config
		rulesFS hackpadfs.FS
		mac     string
	}
	var tests []struct {
		name   string
		fields fields
		want   EvaluationResult
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RegoEvaluator{
				config:  tt.fields.config,
				rulesFS: tt.fields.rulesFS,
				mac:     tt.fields.mac,
			}
			if got, _ := r.EvaluateExpressions(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EvaluateExpressions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegoEvaluator_evaluateExpression(t *testing.T) {
	testFS, err := mem.NewFS()
	if err != nil {
		t.Fatal("Failed to setup memfs")
	}
	if writeMemFile(testFS, "alwaysTrue.rego", alwaysTrueRego, t) != nil {
		t.Fatal(err)
	}

	if writeMemFile(testFS, "alwaysFalse.rego", alwaysFalseRego, t) != nil {
		t.Fatal(err)
	}

	check100Percent := `package upswake
default wake = false
wake = true {
	input[i].Name == "cyberpower900"
	input[i].Variables[j].Name == "battery.charge"
	input[i].Variables[j].Value == 100
}`
	if writeMemFile(testFS, "check100Percent.rego", []byte(check100Percent), t) != nil {
		t.Fatal(err)
	}

	type fields struct {
		rulesFS hackpadfs.FS
	}
	type args struct {
		target *entity.TargetServer
		//nutServer *entity.NutServer
		inputJSON string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "nothing to evaluate",
			fields: fields{
				rulesFS: tempFS,
			},
			args: args{
				target:    nil,
				inputJSON: "",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "evaluate with always true rule",
			fields: fields{
				rulesFS: testFS,
			},
			args: args{
				target: &entity.TargetServer{
					Rules: []string{
						"alwaysTrue.rego",
					},
				},
				inputJSON: validNUTOutput, // We don't care about the input JSON, as the rule will always return true for this fs
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "evaluate with always false rule",
			fields: fields{
				rulesFS: testFS,
			},
			args: args{
				target: &entity.TargetServer{
					Rules: []string{
						"alwaysFalse.rego",
					},
				},
				inputJSON: validNUTOutput, // We don't care about the input JSON, as the rule will always return true for this fs
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "file not found",
			fields: fields{
				rulesFS: testFS,
			},
			args: args{
				target: &entity.TargetServer{
					Rules: []string{
						"doesnotexist.rego",
					},
				},
				inputJSON: validNUTOutput, // We don't care about the input JSON, as the rule will always return true for this fs
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "ups 100% check positive",
			fields: fields{
				rulesFS: testFS,
			},
			args: args{
				target: &entity.TargetServer{
					Rules: []string{
						"check100Percent.rego",
					},
				},
				inputJSON: `[{"Name":"cyberpower900","Description":"Unavailable","Master":false,"NumberOfLogins":0,"Clients":[],"Variables":[{"Name":"battery.charge","Value":100,"Type":"INTEGER","Description":"Battery charge (percent of full)","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"},{"Name":"ups.status","Value":"OL","Type":"STRING","Description":"UPS status","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"}]}]`,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "ups 100% check negative",
			fields: fields{
				rulesFS: testFS,
			},
			args: args{
				target: &entity.TargetServer{
					Port: entity.DefaultWoLPort,
					Rules: []string{
						"check100Percent.rego",
					},
				},
				inputJSON: `[{"Name":"cyberpower900","Description":"Unavailable","Master":false,"NumberOfLogins":0,"Clients":[],"Variables":[{"Name":"battery.charge","Value":10,"Type":"INTEGER","Description":"Battery charge (percent of full)","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"},{"Name":"ups.status","Value":"OL","Type":"STRING","Description":"UPS status","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"}]}]`,
			},
			want:    false,
			wantErr: false,
		},
		// TODO: Add more rules that tests inputJSON
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RegoEvaluator{
				rulesFS: tt.fields.rulesFS,
			}
			got, err := r.evaluateExpression(tt.args.target, tt.args.inputJSON)
			if (err != nil) != tt.wantErr {
				t.Errorf("evaluateExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("evaluateExpression() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func writeMemFile(fs hackpadfs.FS, fileName string, contents []byte, t *testing.T) error {
	create, err := hackpadfs.Create(fs, fileName)
	if err != nil {
		t.Fatalf("failed to create memfs file: %s", err)
	}
	_, err = hackpadfs.WriteFile(create, contents)
	if err != nil {
		t.Fatalf("failed to write file %s", err)
	}
	return err
}

func TestRegoEvaluator_evaluateExpressions(t *testing.T) {
	alwaysTrueRegoFS, err := mem.NewFS()
	if err != nil {
		t.Fatal("Failed to setup memfs")
	}
	if writeMemFile(alwaysTrueRegoFS, "test.rego", alwaysTrueRego, t) != nil {
		t.Fatal(err)
	}

	type fields struct {
		config  *entity.Config
		rulesFS hackpadfs.FS
		mac     string
	}

	type args struct {
		getUPSJSON func(server *entity.NutServer) (string, error)
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    EvaluationResult
		wantErr bool
	}{
		{
			name: "valid eval",
			fields: fields{
				config: &entity.Config{
					NutServers: []entity.NutServer{
						{
							Name:     "test",
							Host:     "",
							Port:     entity.DefaultNUTServerPort,
							Username: "",
							Password: "",
							Targets: []entity.TargetServer{
								{
									Name:      "test server",
									MAC:       "00:11:22:33:44:55",
									Broadcast: "192.168.1.255",
									Port:      entity.DefaultWoLPort,
									Interval:  "15m",
									Rules: []string{
										"test.rego",
									},
								},
							},
						},
					}},
				rulesFS: alwaysTrueRegoFS,
				mac:     "00:11:22:33:44:55",
			},
			args: args{
				getUPSJSON: func(server *entity.NutServer) (string, error) { return validNUTOutput, nil },
			},
			want: EvaluationResult{
				Allowed: true,
				Found:   true,
				Target: &entity.TargetServer{
					Name:      "test server",
					MAC:       "00:11:22:33:44:55",
					Broadcast: "192.168.1.255",
					Port:      entity.DefaultWoLPort,
					Interval:  "15m",
					Rules: []string{
						"test.rego",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing mac in config",
			fields: fields{
				config: &entity.Config{
					NutServers: []entity.NutServer{
						{
							Name:     "test",
							Host:     "",
							Port:     entity.DefaultNUTServerPort,
							Username: "",
							Password: "",
							Targets: []entity.TargetServer{
								{
									Name:      "test server",
									MAC:       "00:00:00:00:00:00",
									Broadcast: "192.168.1.255",
									Port:      entity.DefaultWoLPort,
									Interval:  "15m",
									Rules: []string{
										"test.rego",
									},
								},
							},
						},
					}},
				rulesFS: alwaysTrueRegoFS,
				mac:     "00:11:22:33:44:55",
			},
			args: args{
				getUPSJSON: func(server *entity.NutServer) (string, error) { return validNUTOutput, nil },
			},
			want: EvaluationResult{
				Allowed: false,
				Found:   false,
				Target:  nil,
			},
			wantErr: false,
		},
		{
			name: "invalid nutserver output",
			fields: fields{
				config: &entity.Config{
					NutServers: []entity.NutServer{
						{
							Name:     "test",
							Host:     "",
							Port:     entity.DefaultNUTServerPort,
							Username: "",
							Password: "",
							Targets: []entity.TargetServer{
								{
									Name:      "test server",
									MAC:       "00:11:22:33:44:55",
									Broadcast: "192.168.1.255",
									Port:      entity.DefaultWoLPort,
									Interval:  "15m",
									Rules: []string{
										"test.rego",
									},
								},
							},
						},
					}},
				rulesFS: alwaysTrueRegoFS,
				mac:     "00:11:22:33:44:55",
			},
			args: args{
				getUPSJSON: func(server *entity.NutServer) (string, error) { return invalidNUTOutput, nil },
			},
			want: EvaluationResult{
				Allowed: false,
				Found:   true,
				Target: &entity.TargetServer{
					Name:      "test server",
					MAC:       "00:11:22:33:44:55",
					Broadcast: "192.168.1.255",
					Port:      entity.DefaultWoLPort,
					Interval:  "15m",
					Rules: []string{
						"test.rego",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RegoEvaluator{
				config:  tt.fields.config,
				rulesFS: tt.fields.rulesFS,
				mac:     tt.fields.mac,
			}
			got, err := r.evaluateExpressions(tt.args.getUPSJSON)
			if (err != nil) != tt.wantErr {
				t.Errorf("evaluateExpressions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("evaluateExpressions() got = %v, want %v", got, tt.want)
			}
		})
	}
}
