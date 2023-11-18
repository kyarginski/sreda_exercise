package config

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestMustLoad(t *testing.T) {
	err := os.Setenv("SENDER_CONFIG_PATH", "../../config/prod.yaml")
	assert.NoError(t, err)
	tests := []struct {
		name string
		want *Config
	}{
		{
			name: "Good case",
			want: &Config{
				Env:     "prod",
				Version: "1.0.0",
				URL:     "http://localhost:8091",
				Requests: struct {
					Amount    int64 `yaml:"amount"`
					PerSecond int64 `yaml:"per_second"`
				}{
					Amount:    1000,
					PerSecond: 10,
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := MustLoad()
			diff := cmp.Diff(tt.want, got)
			if diff != "" {
				t.Errorf("MustLoad() = %v, want %v", got, tt.want)
			}
		})
	}
}
