package repository

import (
	"fmt"
	"strings"

	"inventory/backend/internal/domain"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Searchable represents an interface for models that can be indexed for global search.
type Searchable interface {
	GetID() uint
	GetSearchableContent() string
	GetEntityType() string
}

// UpdateSearchIndex updates the GlobalSearchEntry for a given searchable entity.
// It should be called within a transaction when the entity is created or updated.
func UpdateSearchIndex(tx *gorm.DB, entity Searchable) error {
	entityType := entity.GetEntityType()
	entityID := entity.GetID()
	content := entity.GetSearchableContent()

	// Coalesce and convert content to tsvector
	// We use raw SQL because GORM doesn't natively support tsvector functions well.
	// The query builds a tsvector from the provided content string.
	vectorQuery := `to_tsvector('english', ?)`

	// Use ON CONFLICT to perform an "upsert".
	// If a search entry for this entity already exists, update it. Otherwise, create it.
	upsertSQL := `
        INSERT INTO global_search_entries (entity_type, entity_id, content, search_vector, created_at, updated_at)
        VALUES (?, ?, ?, (%s), NOW(), NOW())
        ON CONFLICT (entity_type, entity_id) DO UPDATE SET
            content = EXCLUDED.content,
            search_vector = EXCLUDED.search_vector,
            updated_at = NOW();
    `
	fullSQL := fmt.Sprintf(upsertSQL, vectorQuery)

	return tx.Exec(fullSQL, entityType, entityID, content, content).Error
}

// DeleteFromSearchIndex removes a GlobalSearchEntry for a given entity.
// It should be called within a transaction when the entity is deleted.
func DeleteFromSearchIndex(tx *gorm.DB, entityType string, entityID uint) error {
	return tx.Where("entity_type = ? AND entity_id = ?", entityType, entityID).Delete(&domain.GlobalSearchEntry{}).Error
}



// SearchResult defines the structure for items returned by the global search.
type SearchResult struct {
	EntityType string      `json:"entity_type"`
	EntityID   uint        `json:"entity_id"`
	Content    string      `json:"content"`
	Rank       float32     `json:"-"` // For internal sorting, not exposed
	Entity     interface{} `json:"entity"`
}

// SearchRepository performs search operations.
type SearchRepository struct {
	db *gorm.DB
}

// NewSearchRepository creates a new search repository.
func NewSearchRepository(db *gorm.DB) *SearchRepository {
	return &SearchRepository{db: db}
}

// Search performs a full-text search across all indexed entities.
func (r *SearchRepository) Search(query string) ([]SearchResult, error) {
	var results []SearchResult

	// Process the search query to make it a valid tsquery.
	// Replace spaces with '&' for "AND" logic. You can make this more complex (e.g., support OR, prefixes).
	tsQuery := strings.ReplaceAll(strings.TrimSpace(query), " ", "&")

	// This query searches the global_search_entries table.
	// It uses `to_tsquery` to match against the `search_vector`.
	// `ts_rank` is used to score the relevance of each match.
	// Results are ordered by this rank.
	err := r.db.Raw(`
        SELECT 
            entity_type, 
            entity_id, 
            content,
            ts_rank(search_vector, to_tsquery('english', ?)) as rank
        FROM 
            global_search_entries
        WHERE 
            search_vector @@ to_tsquery('english', ?)
        ORDER BY 
            rank DESC
        LIMIT 50;
    `, tsQuery, tsQuery).Scan(&results).Error

	if err != nil {
		logrus.Errorf("Error executing search query: %v", err)
		return nil, err
	}

	return results, nil
}
