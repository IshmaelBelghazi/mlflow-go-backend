package sql

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/mlflow/mlflow-go-backend/pkg/tracking/store/sql/models"
)

func TestDottedSearchUsesBoundKeys(t *testing.T) {
	t.Parallel()

	factories := []func() gorm.Dialector{
		newPostgresDialector, newSqliteDialector, newSQLServerDialector, newMySQLDialector,
	}
	for _, factory := range factories {
		database, err := gorm.Open(factory(), &gorm.Config{DryRun: true})
		require.NoError(t, err)
		t.Run(database.Dialector.Name(), func(t *testing.T) {
			var previous *gorm.Statement
			for _, filter := range []string{"tags.`life.worker` = 'Trainer'", "tags.life.worker = 'Trainer'"} {
				query := database.Model(&models.Run{})
				require.Nil(t, applyFilter(context.Background(), database, query, filter))
				require.NoError(t, query.Select("runs.run_uuid").Find(&models.Run{}).Error)
				assert.Equal(t, []any{"life.worker", "Trainer"}, query.Statement.Vars)
				if previous != nil {
					assert.Equal(t, previous.SQL.String(), query.Statement.SQL.String())
				}
				previous = query.Statement
			}
		})
	}
}

func TestDottedSearchMatchesExactStoredKeys(t *testing.T) {
	t.Parallel()

	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	connection, err := database.DB()
	require.NoError(t, err)
	connection.SetMaxOpenConns(1)
	t.Cleanup(func() { assert.NoError(t, connection.Close()) })

	for _, statement := range []string{
		"CREATE TABLE runs (run_uuid TEXT PRIMARY KEY)",
		"CREATE TABLE tags (run_uuid TEXT, key TEXT, value TEXT)",
		"INSERT INTO runs VALUES ('match'), ('short-key'), ('missing')",
		"INSERT INTO tags VALUES ('match', 'life.worker', 'Trainer')",
		"INSERT INTO tags VALUES ('short-key', 'life', 'Trainer'), ('short-key', 'worker', 'Trainer')",
	} {
		require.NoError(t, database.Exec(statement).Error)
	}

	cases := []struct {
		filter string
		want   []string
	}{
		{"tags.life.worker = 'Trainer'", []string{"match"}},
		{"tags.`life.worker` = 'Trainer'", []string{"match"}},
		{"tags.life.worker IS NOT NULL", []string{"match"}},
		{"tags.life.worker IS NULL", []string{"missing", "short-key"}},
		{"tags.life.worker = 'Trainer' AND tags.worker IS NULL", []string{"match"}},
	}
	for _, sample := range cases {
		t.Run(sample.filter, func(t *testing.T) {
			query := database.Model(&models.Run{})
			require.Nil(t, applyFilter(context.Background(), database, query, sample.filter))
			var ids []string
			require.NoError(t, query.Order("runs.run_uuid").Pluck("runs.run_uuid", &ids).Error)
			assert.Equal(t, sample.want, ids)
		})
	}
}
