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
	validNUTOutput = "[{\"Name\":\"cyberpower900\",\"Description\":\"Unavailable\",\"Master\":false,\"NumberOfLogins\":1,\"Clients\":[\"127.0.0.1\"],\"Variables\":[{\"Name\":\"battery.charge\",\"Value\":100,\"Type\":\"INTEGER\",\"Description\":\"Battery charge (percent of full)\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"battery.charge.low\",\"Value\":10,\"Type\":\"INTEGER\",\"Description\":\"Remaining battery level when UPS switches to LB (percent)\",\"Writeable\":true,\"MaximumLength\":10,\"OriginalType\":\"STRING\"},{\"Name\":\"battery.charge.warning\",\"Value\":20,\"Type\":\"INTEGER\",\"Description\":\"Battery level when UPS switches to Warning state (percent)\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"battery.mfr.date\",\"Value\":\"CPS\",\"Type\":\"STRING\",\"Description\":\"Battery manufacturing date\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"battery.runtime\",\"Value\":2820,\"Type\":\"INTEGER\",\"Description\":\"Battery runtime (seconds)\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"battery.runtime.low\",\"Value\":300,\"Type\":\"INTEGER\",\"Description\":\"Remaining battery runtime when UPS switches to LB (seconds)\",\"Writeable\":true,\"MaximumLength\":10,\"OriginalType\":\"STRING\"},{\"Name\":\"battery.type\",\"Value\":\"PbAcid\",\"Type\":\"STRING\",\"Description\":\"Battery chemistry\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"battery.voltage\",\"Value\":24,\"Type\":\"FLOAT_64\",\"Description\":\"Battery voltage (V)\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"battery.voltage.nominal\",\"Value\":24,\"Type\":\"INTEGER\",\"Description\":\"Nominal battery voltage (V)\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"device.mfr\",\"Value\":\"CPS\",\"Type\":\"STRING\",\"Description\":\"Description unavailable\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"device.model\",\"Value\":\"CP900EPFCLCD\",\"Type\":\"STRING\",\"Description\":\"Description unavailable\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"device.serial\",\"Value\":0,\"Type\":\"INTEGER\",\"Description\":\"Description unavailable\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"device.type\",\"Value\":\"ups\",\"Type\":\"STRING\",\"Description\":\"Description unavailable\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"driver.name\",\"Value\":\"usbhid-ups\",\"Type\":\"STRING\",\"Description\":\"Driver name\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"driver.parameter.bus\",\"Value\":1,\"Type\":\"INTEGER\",\"Description\":\"Description unavailable\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"driver.parameter.pollfreq\",\"Value\":30,\"Type\":\"INTEGER\",\"Description\":\"Description unavailable\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"driver.parameter.pollinterval\",\"Value\":15,\"Type\":\"INTEGER\",\"Description\":\"Description unavailable\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"driver.parameter.port\",\"Value\":\"auto\",\"Type\":\"STRING\",\"Description\":\"Description unavailable\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"driver.parameter.product\",\"Value\":\"CP900EPFCLCD\",\"Type\":\"STRING\",\"Description\":\"Description unavailable\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"driver.parameter.productid\",\"Value\":501,\"Type\":\"INTEGER\",\"Description\":\"Description unavailable\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"driver.parameter.serial\",\"Value\":0,\"Type\":\"INTEGER\",\"Description\":\"Description unavailable\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"driver.parameter.synchronous\",\"Value\":\"no\",\"Type\":\"STRING\",\"Description\":\"Description unavailable\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"driver.parameter.vendor\",\"Value\":\"CPS\",\"Type\":\"STRING\",\"Description\":\"Description unavailable\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"driver.parameter.vendorid\",\"Value\":764,\"Type\":\"INTEGER\",\"Description\":\"Description unavailable\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"driver.version\",\"Value\":\"2.7.4\",\"Type\":\"NUMBER\",\"Description\":\"Driver version - NUT release\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"\"},{\"Name\":\"driver.version.data\",\"Value\":\"CyberPower HID 0.4\",\"Type\":\"STRING\",\"Description\":\"Description unavailable\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"driver.version.internal\",\"Value\":0.41,\"Type\":\"FLOAT_64\",\"Description\":\"Internal driver version\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"input.transfer.high\",\"Value\":260,\"Type\":\"INTEGER\",\"Description\":\"High voltage transfer point (V)\",\"Writeable\":true,\"MaximumLength\":10,\"OriginalType\":\"STRING\"},{\"Name\":\"input.transfer.low\",\"Value\":170,\"Type\":\"INTEGER\",\"Description\":\"Low voltage transfer point (V)\",\"Writeable\":true,\"MaximumLength\":10,\"OriginalType\":\"STRING\"},{\"Name\":\"input.voltage\",\"Value\":241,\"Type\":\"FLOAT_64\",\"Description\":\"Input voltage (V)\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"input.voltage.nominal\",\"Value\":230,\"Type\":\"INTEGER\",\"Description\":\"Nominal input voltage (V)\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"output.voltage\",\"Value\":260,\"Type\":\"FLOAT_64\",\"Description\":\"Output voltage (V)\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"ups.beeper.status\",\"Value\":true,\"Type\":\"STRING\",\"Description\":\"UPS beeper status\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"ups.delay.shutdown\",\"Value\":20,\"Type\":\"INTEGER\",\"Description\":\"Interval to wait after shutdown with delay command (seconds)\",\"Writeable\":true,\"MaximumLength\":10,\"OriginalType\":\"STRING\"},{\"Name\":\"ups.delay.start\",\"Value\":30,\"Type\":\"INTEGER\",\"Description\":\"Interval to wait before (re)starting the load (seconds)\",\"Writeable\":true,\"MaximumLength\":10,\"OriginalType\":\"STRING\"},{\"Name\":\"ups.load\",\"Value\":10,\"Type\":\"INTEGER\",\"Description\":\"Load on UPS (percent of full)\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"ups.mfr\",\"Value\":\"CPS\",\"Type\":\"STRING\",\"Description\":\"UPS manufacturer\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"ups.model\",\"Value\":\"CP900EPFCLCD\",\"Type\":\"STRING\",\"Description\":\"UPS model\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"ups.productid\",\"Value\":501,\"Type\":\"INTEGER\",\"Description\":\"Product ID for USB devices\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"ups.realpower.nominal\",\"Value\":540,\"Type\":\"INTEGER\",\"Description\":\"UPS real power rating (W)\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"ups.serial\",\"Value\":0,\"Type\":\"INTEGER\",\"Description\":\"UPS serial number\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"ups.status\",\"Value\":\"OL\",\"Type\":\"STRING\",\"Description\":\"UPS status\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"ups.test.result\",\"Value\":\"No test initiated\",\"Type\":\"STRING\",\"Description\":\"Results of last self test\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"ups.timer.shutdown\",\"Value\":-60,\"Type\":\"INTEGER\",\"Description\":\"Time before the load will be shutdown (seconds)\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"ups.timer.start\",\"Value\":-60,\"Type\":\"INTEGER\",\"Description\":\"Time before the load will be started (seconds)\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"},{\"Name\":\"ups.vendorid\",\"Value\":764,\"Type\":\"INTEGER\",\"Description\":\"Vendor ID for USB devices\",\"Writeable\":false,\"MaximumLength\":0,\"OriginalType\":\"NUMBER\"}],\"Commands\":[{\"Name\":\"beeper.disable\",\"Description\":\"Disable the UPS beeper\"},{\"Name\":\"beeper.enable\",\"Description\":\"Enable the UPS beeper\"},{\"Name\":\"beeper.mute\",\"Description\":\"Temporarily mute the UPS beeper\"},{\"Name\":\"beeper.off\",\"Description\":\"Obsolete (use beeper.disable or beeper.mute)\"},{\"Name\":\"beeper.on\",\"Description\":\"Obsolete (use beeper.enable)\"},{\"Name\":\"load.off\",\"Description\":\"Turn off the load immediately\"},{\"Name\":\"load.off.delay\",\"Description\":\"Turn off the load with a delay (seconds)\"},{\"Name\":\"load.on\",\"Description\":\"Turn on the load immediately\"},{\"Name\":\"load.on.delay\",\"Description\":\"Turn on the load with a delay (seconds)\"},{\"Name\":\"shutdown.return\",\"Description\":\"Turn off the load and return when power is back\"},{\"Name\":\"shutdown.stayoff\",\"Description\":\"Turn off the load and remain off\"},{\"Name\":\"shutdown.stop\",\"Description\":\"Stop a shutdown in progress\"},{\"Name\":\"test.battery.start.deep\",\"Description\":\"Start a deep battery test\"},{\"Name\":\"test.battery.start.quick\",\"Description\":\"Start a quick battery test\"},{\"Name\":\"test.battery.stop\",\"Description\":\"Stop the battery test\"}]}]"
	emptyNUTOutput = ""
)

var (
	defaultConfig, _ = viper.CreateDefaultConfig()
	tempFS, _        = mem.NewFS()
	alwaysTrueRego   = []byte("package upswake\n\ndefault wake = true")
	alwaysFalseRego  = []byte("package upswake\n\ndefault wake = false")
)

func TestNewRegoEvaluator(t *testing.T) {
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
			name: "valid config",
			args: args{
				config:  defaultConfig,
				mac:     "00:00:00:00:00:00",
				rulesFS: tempFS,
			},
			want: &RegoEvaluator{
				config:  defaultConfig,
				rulesFS: tempFS,
				mac:     "00:00:00:00:00:00",
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
	type fields struct {
		config  *entity.Config
		rulesFS hackpadfs.FS
		mac     string
	}
	type args struct {
		target    *entity.TargetServer
		nutServer *entity.NutServer
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
				config:  defaultConfig,
				rulesFS: tempFS,
				mac:     "00:00:00:00:00:00",
			},
			args: args{
				target:    nil,
				nutServer: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RegoEvaluator{
				config:  tt.fields.config,
				rulesFS: tt.fields.rulesFS,
				mac:     tt.fields.mac,
			}
			got, err := r.evaluateExpression(tt.args.target, tt.args.nutServer, validNUTOutput)
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
				getUPSJSON: func(server *entity.NutServer) (string, error) { return emptyNUTOutput, nil },
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
