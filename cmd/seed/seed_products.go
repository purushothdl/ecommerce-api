// cmd/seed/seed_products.go
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/lib/pq"
	"github.com/purushothdl/ecommerce-api/internal/models"
	apperrors "github.com/purushothdl/ecommerce-api/pkg/errors"

)

// Define structs to match the JSON structure from the dummy API
type APIProductResponse struct {
	Products []APIProduct `json:"products"`
	Total    int          `json:"total"`
	Skip     int          `json:"skip"`
	Limit    int          `json:"limit"`
}

type APIProduct struct {
	ID                  int         `json:"id"`
	Title               string      `json:"title"`
	Description         string      `json:"description"`
	Category            string      `json:"category"`
	Price               float64     `json:"price"`
	Stock               int         `json:"stock"`
	Brand               string      `json:"brand"`
	SKU                 string      `json:"sku"`
	Images              []string    `json:"images"`
	Thumbnail           string      `json:"thumbnail"`
	Dimensions          interface{} `json:"dimensions"` 
	WarrantyInformation string      `json:"warrantyInformation"`
}

var SeedProductsTask = SeederTask{
	Name: "Seed Products",
	Run: func(ctx context.Context, deps SeederDeps) error {
		log.Println("Fetching products from dummyjson API...")
		resp, err := http.Get("https://dummyjson.com/products?limit=200")
		if err != nil {
			return fmt.Errorf("failed to fetch products: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		var apiResponse APIProductResponse
		if err := json.Unmarshal(body, &apiResponse); err != nil {
			return fmt.Errorf("failed to unmarshal products JSON: %w", err)
		}

		log.Printf("Fetched %d products. Seeding to database...", len(apiResponse.Products))

		// Use a map to cache category IDs to avoid redundant DB calls
		categoryCache := make(map[string]*models.Category)

		for _, apiProduct := range apiResponse.Products {
			// 1. Get or Create Category
			categoryName := strings.ReplaceAll(strings.Title(apiProduct.Category), "-", " ")
			category, exists := categoryCache[categoryName]
			if !exists {
				// Try to get from DB first
				dbCat, err := deps.CategoryRepo.GetByName(ctx, categoryName)
				if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
					log.Printf("Warning: could not get category %s: %v. Skipping product.", categoryName, err)
					continue
				}
				if dbCat != nil {
					category = dbCat
				} else {
					// Create if not in DB
					newCat := &models.Category{Name: categoryName}
					if err := deps.CategoryRepo.Create(ctx, newCat); err != nil {
						log.Printf("Warning: could not create category %s: %v. Skipping product.", categoryName, err)
						continue
					}
					category = newCat
				}
				categoryCache[categoryName] = category
			}

			// 2. Marshal dimensions to JSONB
			dimensionsJSON, err := json.Marshal(apiProduct.Dimensions)
			if err != nil {
				log.Printf("Warning: could not marshal dimensions for product %s: %v. Skipping.", apiProduct.Title, err)
				continue
			}

			// 3. Create Product Model
			product := &models.Product{
				Name:                apiProduct.Title,
				Description:         apiProduct.Description,
				Price:               apiProduct.Price,
				StockQuantity:       apiProduct.Stock,
				CategoryID:          category.ID,
				Brand:               apiProduct.Brand,
				SKU:                 apiProduct.SKU,
				Images:              pq.StringArray(apiProduct.Images),
				Thumbnail:           apiProduct.Thumbnail,
				Dimensions:          dimensionsJSON,
				WarrantyInformation: apiProduct.WarrantyInformation,
			}

			// 4. Insert Product into DB
			if err := deps.ProductRepo.Create(ctx, product); err != nil {
				log.Printf("Warning: failed to insert product '%s' (SKU: %s): %v", product.Name, product.SKU, err)
				// Continue to the next product instead of failing the whole seed
			}
		}

		log.Println("Product seeding process finished.")
		return nil
	},
}