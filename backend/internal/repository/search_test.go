//go:build integration

package repository

import (
	"os"
	"testing"

	"inventory/backend/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestSearchRepository_HandlesSpecialChars(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is required")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE IF NOT EXISTS global_search_entries (
			entity_type TEXT NOT NULL,
			entity_id BIGINT NOT NULL,
			content TEXT NOT NULL,
			search_vector TSVECTOR NOT NULL,
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL,
			PRIMARY KEY (entity_type, entity_id)
		)
	`).Error)
	t.Cleanup(func() {
		db.Where("entity_type = ? AND entity_id = ?", "product", 1).Delete(&domain.GlobalSearchEntry{})
	})

	require.NoError(t, UpdateSearchIndex(db, &searchTestEntity{}))
	repo := NewSearchRepository(db)

	for _, query := range []string{"coffee:", "tea&cake", "a:b:c", "x & y"} {
		_, err := repo.Search(query)
		assert.NoError(t, err, "query %q must not error", query)
	}
}

type searchTestEntity struct{}

func (*searchTestEntity) GetID() uint                  { return 1 }
func (*searchTestEntity) GetSearchableContent() string { return "Coffee Mug" }
func (*searchTestEntity) GetEntityType() string        { return "product" }

var _ Searchable = (*searchTestEntity)(nil)
