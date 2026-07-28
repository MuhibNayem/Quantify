package repository

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type reportsRepository struct {
	DB *gorm.DB
}

func NewReportsRepository(db *gorm.DB) ReportsRepository {
	return &reportsRepository{DB: db}
}

func parseFlexibleTime(s string) (time.Time, error) {
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized timestamp: %q", s)
}

func isValidTimezone(tz string) bool {
	if tz == "" {
		return false
	}
	for _, r := range tz {
		switch {
		case r >= 'A' && r <= 'Z':
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '/' || r == '_' || r == '+' || r == '-':
		default:
			return false
		}
	}
	return true
}
