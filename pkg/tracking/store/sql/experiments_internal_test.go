package sql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/mlflow/mlflow-go-backend/pkg/tracking/store/sql/models"
)

func TestApplyExperimentsFilterSupportsTagIsNull(t *testing.T) {
	t.Parallel()

	database, err := gorm.Open(newPostgresDialector(), &gorm.Config{DryRun: true})
	require.NoError(t, err)

	query := database.Model(&models.Experiment{})
	query, contractErr := applyExperimentsFilter(
		database,
		query,
		"tags.`mlflow.experiment.isGateway` IS NULL",
	)

	require.Nil(t, contractErr)
	require.NoError(t, query.Find(&models.Experiment{}).Error)

	expectedSQL := `
	SELECT * FROM "experiments"
	WHERE NOT EXISTS (
		SELECT 1 FROM "experiment_tags"
		WHERE key = $1
		AND experiment_tags.experiment_id = experiments.experiment_id
	)`

	assert.Equal(t, removeWhitespace(expectedSQL), removeWhitespace(query.Statement.SQL.String()))
	assert.Equal(t, []any{"mlflow.experiment.isGateway"}, query.Statement.Vars)
}
