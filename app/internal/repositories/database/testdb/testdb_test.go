package testdb_test

import (
	"context"
	"testing"

	"github.com/okocraft/auth-service/internal/repositories/database"
	"github.com/okocraft/auth-service/internal/repositories/database/testdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTestDB(t *testing.T) {
	tests := []struct {
		name  string
		useTx bool
	}{
		{
			name:  "success: useTx=true",
			useTx: true,
		},
		{
			name:  "success: useTx=false",
			useTx: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()

			db, err := testdb.NewTestDB(tt.useTx)
			require.NoError(t, err)
			require.NotNil(t, db)

			t.Cleanup(func() {
				require.NoError(t, db.Cleanup())
				assert.Error(t, db.GetDB().Base().PingContext(ctx), "connection is still alive")
			})

			runCount := 0
			db.Run(t, func(ctx context.Context, conn database.Connection) {
				assert.NotNil(t, ctx)
				assert.NotNil(t, conn)
				runCount++
			})
			assert.Equal(t, 1, runCount, "Run function was called more than once, or not called")
			assert.NoError(t, db.GetDB().Base().PingContext(ctx))
		})
	}
}
