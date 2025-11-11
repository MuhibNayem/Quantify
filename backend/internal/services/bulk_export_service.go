package services

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"inventory/backend/internal/domain"

	"github.com/xuri/excelize/v2"
)

type BulkExportService struct{}

func NewBulkExportService() *BulkExportService {
	return &BulkExportService{}
}

func (s *BulkExportService) GenerateProductExport(products []domain.Product, format string) (*bytes.Buffer, error) {
	switch format {
	case "csv":
		return s.generateProductCSV(products)
	case "excel":
		return s.generateProductExcel(products)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

func (s *BulkExportService) generateProductCSV(products []domain.Product) (*bytes.Buffer, error) {
	var b bytes.Buffer
	writer := csv.NewWriter(&b)

	header := []string{"ID", "SKU", "Name", "Description", "Category", "SubCategory", "Supplier", "Brand", "PurchasePrice", "SellingPrice", "BarcodeUPC", "Status"}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	for _, p := range products {
		row := []string{
			fmt.Sprintf("%d", p.ID),
			p.SKU,
			p.Name,
			p.Description,
			p.Category.Name,
			p.SubCategory.Name,
			p.Supplier.Name,
			p.Brand,
			fmt.Sprintf("%.2f", p.PurchasePrice),
			fmt.Sprintf("%.2f", p.SellingPrice),
			p.BarcodeUPC,
			p.Status,
		}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return &b, nil
}

func (s *BulkExportService) generateProductExcel(products []domain.Product) (*bytes.Buffer, error) {
	f := excelize.NewFile()
	sheetName := "Products"
	index, _ := f.NewSheet(sheetName)

	header := []string{"ID", "SKU", "Name", "Description", "Category", "SubCategory", "Supplier", "Brand", "PurchasePrice", "SellingPrice", "BarcodeUPC", "Status"}
	for i, h := range header {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, h)
	}

	for i, p := range products {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), p.ID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), p.SKU)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), p.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), p.Description)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), p.Category.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), p.SubCategory.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), p.Supplier.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), p.Brand)
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), p.PurchasePrice)
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), p.SellingPrice)
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), p.BarcodeUPC)
		f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), p.Status)
	}

	f.SetActiveSheet(index)

	return f.WriteToBuffer()
}
