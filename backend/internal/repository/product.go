package repository

import (
	"inventory/backend/internal/domain"

	"gorm.io/gorm"
)

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) CreateProduct(product *domain.Product) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Create the product
		if err := tx.Create(product).Error; err != nil {
			return err
		}

		// Update the search index
		if err := UpdateSearchIndex(tx, product); err != nil {
			return err
		}

		return nil
	})
}

func (r *productRepository) UpdateProduct(product *domain.Product, updates map[string]interface{}) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Update the product
		if err := tx.Model(product).Updates(updates).Error; err != nil {
			return err
		}

		// Reload the product to get all fields for indexing
		if err := tx.First(product, product.ID).Error; err != nil {
			return err
		}

		// Update the search index
		if err := UpdateSearchIndex(tx, product); err != nil {
			return err
		}

		return nil
	})
}

func (r *productRepository) DeleteProduct(product *domain.Product) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete the product
		if err := tx.Delete(product).Error; err != nil {
			return err
		}

		// Delete from the search index
		if err := DeleteFromSearchIndex(tx, product.GetEntityType(), product.GetID()); err != nil {
			return err
		}

		return nil
	})
}

func (r *productRepository) GetProductBySKU(sku string) (*domain.Product, error) {
	var product domain.Product
	if err := r.db.Preload("Category").Preload("SubCategory").Preload("Supplier").Preload("Location").
		Where("sku = ?", sku).First(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) GetProductByBarcode(barcode string) (*domain.Product, error) {
	var product domain.Product
	if err := r.db.Preload("Category").Preload("SubCategory").Preload("Supplier").Preload("Location").
		Where("barcode_upc = ?", barcode).First(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) CreateBulk(products *[]domain.Product) error {
	return r.db.CreateInBatches(products, 100).Error
}

func (r *productRepository) GetAll() ([]domain.Product, error) {
	var products []domain.Product
	if err := r.db.Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (r *productRepository) GetFiltered(filters map[string]interface{}) ([]domain.Product, error) {
	var products []domain.Product
	query := r.db.Model(&domain.Product{})

	if val, ok := filters["category"]; ok && val != "" {
		query = query.Where("category_id = ?", val)
	}
	if val, ok := filters["supplier"]; ok && val != "" {
		query = query.Where("supplier_id = ?", val)
	}

	if err := query.Preload("Category").Preload("SubCategory").Preload("Supplier").Preload("Location").Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}
