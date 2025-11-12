package migrations

import (
	"inventory/backend/internal/domain"
	"inventory/backend/internal/repository"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// PopulateSearchIndex populates the global_search_entries table with all existing data.
// It first truncates the table to ensure a clean state.
func PopulateSearchIndex(db *gorm.DB) error {
	logrus.Info("Starting to populate the global search index...")

	// 1. Truncate the existing search index to prevent duplicates and remove old entries.
	logrus.Info("Truncating global_search_entries table...")
	if err := db.Exec("TRUNCATE TABLE global_search_entries RESTART IDENTITY;").Error; err != nil {
		logrus.Errorf("Failed to truncate global_search_entries table: %v", err)
		return err
	}

	// 2. Index all models
	if err := indexProducts(db); err != nil {
		return err
	}
	if err := indexUsers(db); err != nil {
		return err
	}
	if err := indexSuppliers(db); err != nil {
		return err
	}
	if err := indexCategories(db); err != nil {
		return err
	}

	logrus.Info("Successfully populated the global search index.")
	return nil
}

func indexProducts(db *gorm.DB) error {
	logrus.Info("Indexing all products...")
	var products []domain.Product
	if err := db.FindInBatches(&products, 100, func(tx *gorm.DB, batch int) error {
		for i := range products {
			if err := repository.UpdateSearchIndex(tx, &products[i]); err != nil {
				return err
			}
		}
		logrus.Infof("Indexed batch %d of products", batch)
		return nil
	}).Error; err != nil {
		logrus.Errorf("Failed while indexing products: %v", err)
		return err
	}
	logrus.Info("Finished indexing products.")
	return nil
}

func indexUsers(db *gorm.DB) error {
	logrus.Info("Indexing all users...")
	var users []domain.User
	if err := db.FindInBatches(&users, 100, func(tx *gorm.DB, batch int) error {
		for i := range users {
			if err := repository.UpdateSearchIndex(tx, &users[i]); err != nil {
				return err
			}
		}
		logrus.Infof("Indexed batch %d of users", batch)
		return nil
	}).Error; err != nil {
		logrus.Errorf("Failed while indexing users: %v", err)
		return err
	}
	logrus.Info("Finished indexing users.")
	return nil
}

func indexSuppliers(db *gorm.DB) error {
	logrus.Info("Indexing all suppliers...")
	var suppliers []domain.Supplier
	if err := db.FindInBatches(&suppliers, 100, func(tx *gorm.DB, batch int) error {
		for i := range suppliers {
			if err := repository.UpdateSearchIndex(tx, &suppliers[i]); err != nil {
				return err
			}
		}
		logrus.Infof("Indexed batch %d of suppliers", batch)
		return nil
	}).Error; err != nil {
		logrus.Errorf("Failed while indexing suppliers: %v", err)
		return err
	}
	logrus.Info("Finished indexing suppliers.")
	return nil
}

func indexCategories(db *gorm.DB) error {
	logrus.Info("Indexing all categories...")
	var categories []domain.Category
	if err := db.FindInBatches(&categories, 100, func(tx *gorm.DB, batch int) error {
		for i := range categories {
			if err := repository.UpdateSearchIndex(tx, &categories[i]); err != nil {
				return err
			}
		}
		logrus.Infof("Indexed batch %d of categories", batch)
		return nil
	}).Error; err != nil {
		logrus.Errorf("Failed while indexing categories: %v", err)
		return err
	}
	logrus.Info("Finished indexing categories.")
	return nil
}
