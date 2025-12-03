package testdb_test

import (
	"path/filepath"
	"testing"

	"github.com/okocraft/auth-service/internal/repositories/database/testdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetProjectRoot(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "success",
			want:    "../../../../",
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantDir, err := filepath.Abs(tt.want)
			require.NoError(t, err)
			got, err := testdb.GetProjectRoot()
			if !tt.wantErr(t, err) {
				return
			}
			assert.Equal(t, wantDir, got)
		})
	}
}
