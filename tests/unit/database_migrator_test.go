package unit

import (
	"context"
	"io/fs"
	"sort"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mentor-api/internal/infrastructure/database"
)

func TestMigrationsFS_EmbeddedFiles(t *testing.T) {
	entries, err := fs.ReadDir(database.MigrationsFS, "migrations")
	require.NoError(t, err, "Embedded migrations folder must be readable")

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}

	require.Len(t, files, 2, "Expected exactly 2 migration versions")
	assert.Equal(t, "000001_init_schema.sql", files[0])
	assert.Equal(t, "000002_add_analytics_indexes.sql", files[1])

	// 1. Verify 000001_init_schema.sql content
	initContent, err := fs.ReadFile(database.MigrationsFS, "migrations/000001_init_schema.sql")
	require.NoError(t, err)
	initSQL := string(initContent)
	assert.Contains(t, initSQL, "CREATE TABLE IF NOT EXISTS users")
	assert.Contains(t, initSQL, "CREATE TABLE IF NOT EXISTS questions")
	assert.Contains(t, initSQL, "CREATE TABLE IF NOT EXISTS exams")
	assert.Contains(t, initSQL, "CREATE TABLE IF NOT EXISTS histories")
	assert.Contains(t, initSQL, "idx_histories_user_id")

	// Verify that the separated indexes are NOT in 000001
	assert.NotContains(t, initSQL, "idx_qoe_got_right", "idx_qoe_got_right must be in 000002")
	assert.NotContains(t, initSQL, "idx_histories_finished", "idx_histories_finished must be in 000002")
	assert.NotContains(t, initSQL, "idx_qoe_updated_at", "idx_qoe_updated_at must be in 000002")
	assert.NotContains(t, initSQL, "idx_exams_user_date", "idx_exams_user_date must be in 000002")

	// 2. Verify 000002_add_analytics_indexes.sql content
	indexContent, err := fs.ReadFile(database.MigrationsFS, "migrations/000002_add_analytics_indexes.sql")
	require.NoError(t, err)
	indexSQL := string(indexContent)
	assert.Contains(t, indexSQL, "idx_qoe_got_right")
	assert.Contains(t, indexSQL, "idx_histories_finished")
	assert.Contains(t, indexSQL, "idx_qoe_updated_at")
	assert.Contains(t, indexSQL, "idx_exams_user_date")
}

func TestRunMigrationsFromFS_NilDB(t *testing.T) {
	mapFS := fstest.MapFS{}
	err := database.RunMigrationsFromFS(context.Background(), nil, mapFS, "nonexistent_dir")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection cannot be nil")
}

func TestMigrationVersionSorting(t *testing.T) {
	unsorted := []string{
		"000003_future_feature.sql",
		"000001_init_schema.sql",
		"000002_add_analytics_indexes.sql",
	}

	sort.Strings(unsorted)

	assert.Equal(t, "000001_init_schema.sql", unsorted[0])
	assert.Equal(t, "000002_add_analytics_indexes.sql", unsorted[1])
	assert.Equal(t, "000003_future_feature.sql", unsorted[2])
}
