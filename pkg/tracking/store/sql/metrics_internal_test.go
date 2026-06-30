package sql

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/mlflow/mlflow-go-backend/pkg/entities"
)

func TestLogModelMetricsSkipsPlainRunMetrics(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn:       sqlDB,
		DriverName: "postgres",
	}), &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	store := TrackingSQLStore{}
	contractErr := store.logModelMetricsWithTransaction(db, "run-id", []*entities.Metric{
		{
			Key:       "accuracy",
			Value:     0.9,
			Timestamp: 123,
			Step:      1,
		},
	})

	require.Nil(t, contractErr)
	require.NoError(t, mock.ExpectationsWereMet())
}
