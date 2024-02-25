package webindexer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				Source: "some/source/path",
				Target: "some/target/path",
				SortBy: "name",
				Order:  "asc",
			},
			wantErr: false,
		},
		{
			name:    "missing source",
			config:  Config{Target: "some/target/path", SortBy: "name", Order: "asc"},
			wantErr: true,
			errMsg:  "source is required",
		},
		{
			name:    "missing target",
			config:  Config{Source: "some/source/path", SortBy: "name", Order: "asc"},
			wantErr: true,
			errMsg:  "target is required",
		},
		{
			name:    "invalid sort_by",
			config:  Config{Source: "some/source/path", Target: "some/target/path", SortBy: "invalid", Order: "asc"},
			wantErr: true,
			errMsg:  "sort_by must be one of: last_modified, name, natural_name",
		},
		{
			name:    "invalid order",
			config:  Config{Source: "some/source/path", Target: "some/target/path", SortBy: "name", Order: "invalid"},
			wantErr: true,
			errMsg:  "order must be one of: asc, desc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
