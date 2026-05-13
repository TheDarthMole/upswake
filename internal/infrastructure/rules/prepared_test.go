package rules

import (
	"testing"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/domain/repository"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validJSON = `[{"Name":"cyberpower900","Description":"Unavailable","Master":false,"NumberOfLogins":0,"Clients":[],"Variables":[{"Name":"battery.charge","Value":100,"Type":"INTEGER","Description":"Battery charge (percent of full)","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"}]}]`

// compile time interface checks
var _ repository.RuleRepository = new(PreparedRepository)

func newTestFS(t *testing.T, files map[string][]byte) afero.Fs {
	t.Helper()
	fs := afero.NewMemMapFs()
	for name, content := range files {
		require.NoError(t, afero.WriteFile(fs, name, content, 0o644))
	}
	return fs
}

func TestNewPreparedRepository_Valid(t *testing.T) {
	fs := newTestFS(t, map[string][]byte{
		"alwaysTrue.rego": []byte(`package upswake
default wake := true`),
		"alwaysFalse.rego": []byte(`package upswake
default wake := false`),
	})

	repo, err := NewPreparedRepository(fs)
	require.NoError(t, err)
	assert.Len(t, repo.RuleNames(), 2)
}

func TestNewPreparedRepository_InvalidRule(t *testing.T) {
	fs := newTestFS(t, map[string][]byte{
		"bad.rego": []byte(`package wrongname
default wake := true`),
	})

	_, err := NewPreparedRepository(fs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid rule")
}

func TestPreparedRepository_Evaluate(t *testing.T) {
	fs := newTestFS(t, map[string][]byte{
		"alwaysTrue.rego": []byte(`package upswake
default wake := true`),
		"alwaysFalse.rego": []byte(`package upswake
default wake := false`),
		"check100.rego": []byte(`package upswake
default wake := false
wake if {
	input[i].Variables[j].Name == "battery.charge"
	input[i].Variables[j].Value == 100
}`),
	})

	repo, err := NewPreparedRepository(fs)
	require.NoError(t, err)

	tests := []struct {
		wantErr  error
		name     string
		ruleName string
		json     string
	}{
		{
			name:     "always true",
			ruleName: "alwaysTrue.rego",
			json:     validJSON,
			wantErr:  nil,
		},
		{
			name:     "always false",
			ruleName: "alwaysFalse.rego",
			json:     validJSON,
			wantErr:  entity.ErrEvaluationFalse,
		},
		{
			name:     "check 100 percent positive",
			ruleName: "check100.rego",
			json:     validJSON,
			wantErr:  nil,
		},
		{
			name:     "rule not found",
			ruleName: "nonexistent.rego",
			json:     validJSON,
			wantErr:  ErrRuleNotFound,
		},
		{
			name:     "invalid json",
			ruleName: "alwaysTrue.rego",
			json:     "not json",
			wantErr:  ErrDecodeFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Evaluate(tt.ruleName, tt.json)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
