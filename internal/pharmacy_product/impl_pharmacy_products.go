package pharmacy_product

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type implPharmacyProductsAPI struct{}

func NewPharmacyProductsApi() PharmacyProductsAPI {
	return &implPharmacyProductsAPI{}
}

// GetProducts returns the pharmacy's products filtered by the `include` query
// parameter: "active" (default) hides soft-deleted items, "inactive" returns only
// soft-deleted items, "all" returns everything.
func (o implPharmacyProductsAPI) GetProducts(c *gin.Context) {
	include := c.DefaultQuery("include", "active")
	switch include {
	case "active", "inactive", "all":
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "include must be one of: active, inactive, all",
		})
		return
	}

	updatePharmacyFunc(c, func(_ *gin.Context, pharmacy *Pharmacy) (*Pharmacy, interface{}, int) {
		visible := make([]Product, 0, len(pharmacy.Products))
		for _, p := range pharmacy.Products {
			switch include {
			case "active":
				if p.Active {
					visible = append(visible, p)
				}
			case "inactive":
				if !p.Active {
					visible = append(visible, p)
				}
			case "all":
				visible = append(visible, p)
			}
		}
		return nil, visible, http.StatusOK
	})
}

// GetProduct returns a single product by id, regardless of active flag (so a
// caller that already knows the id can read the document).
func (o implPharmacyProductsAPI) GetProduct(c *gin.Context) {
	updatePharmacyFunc(c, func(_ *gin.Context, pharmacy *Pharmacy) (*Pharmacy, interface{}, int) {
		productId := c.Param("productId")
		if productId == "" {
			return nil, gin.H{"status": http.StatusBadRequest, "message": "Product ID is required"}, http.StatusBadRequest
		}
		idx := slices.IndexFunc(pharmacy.Products, func(p Product) bool { return p.Id == productId })
		if idx < 0 {
			return nil, gin.H{"status": http.StatusNotFound, "message": "Product not found"}, http.StatusNotFound
		}
		return nil, pharmacy.Products[idx], http.StatusOK
	})
}

func (o implPharmacyProductsAPI) CreateProduct(c *gin.Context) {
	updatePharmacyFunc(c, func(_ *gin.Context, pharmacy *Pharmacy) (*Pharmacy, interface{}, int) {
		var product Product
		if err := c.ShouldBindJSON(&product); err != nil {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Invalid request body",
				"error":   err.Error(),
			}, http.StatusBadRequest
		}

		if product.Name == "" {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Product name is required",
			}, http.StatusBadRequest
		}

		if product.Id == "" || product.Id == "@new" {
			product.Id = uuid.NewString()
		}

		conflict := slices.IndexFunc(pharmacy.Products, func(p Product) bool { return p.Id == product.Id })
		if conflict >= 0 {
			return nil, gin.H{
				"status":  http.StatusConflict,
				"message": "Product with this id already exists",
			}, http.StatusConflict
		}

		// New products default to active=true unless the caller explicitly sent active=false.
		// Since Go zero-value for bool is false, treat any new product as active when the
		// caller didn't explicitly mention active. We can't distinguish unset from false
		// without a pointer, so the convention here is "POST always activates".
		product.Active = true

		pharmacy.Products = append(pharmacy.Products, product)
		return pharmacy, product, http.StatusOK
	})
}

func (o implPharmacyProductsAPI) UpdateProduct(c *gin.Context) {
	updatePharmacyFunc(c, func(_ *gin.Context, pharmacy *Pharmacy) (*Pharmacy, interface{}, int) {
		productId := c.Param("productId")
		if productId == "" {
			return nil, gin.H{"status": http.StatusBadRequest, "message": "Product ID is required"}, http.StatusBadRequest
		}

		var incoming Product
		if err := c.ShouldBindJSON(&incoming); err != nil {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Invalid request body",
				"error":   err.Error(),
			}, http.StatusBadRequest
		}

		idx := slices.IndexFunc(pharmacy.Products, func(p Product) bool { return p.Id == productId })
		if idx < 0 {
			return nil, gin.H{"status": http.StatusNotFound, "message": "Product not found"}, http.StatusNotFound
		}

		existing := pharmacy.Products[idx]
		if incoming.Name != "" {
			existing.Name = incoming.Name
		}
		if incoming.Category.Code != "" || incoming.Category.Value != "" {
			existing.Category = incoming.Category
		}
		if incoming.Stock >= 0 {
			existing.Stock = incoming.Stock
		}
		// Active flag is updatable so a user can reactivate a previously soft-deleted product.
		existing.Active = incoming.Active

		pharmacy.Products[idx] = existing
		return pharmacy, existing, http.StatusOK
	})
}

// DeleteProduct performs a soft delete: it flips Active to false but keeps the
// document in pharmacy.Products. The product is then hidden from GetProducts.
func (o implPharmacyProductsAPI) DeleteProduct(c *gin.Context) {
	updatePharmacyFunc(c, func(_ *gin.Context, pharmacy *Pharmacy) (*Pharmacy, interface{}, int) {
		productId := c.Param("productId")
		if productId == "" {
			return nil, gin.H{"status": http.StatusBadRequest, "message": "Product ID is required"}, http.StatusBadRequest
		}

		idx := slices.IndexFunc(pharmacy.Products, func(p Product) bool { return p.Id == productId })
		if idx < 0 {
			return nil, gin.H{"status": http.StatusNotFound, "message": "Product not found"}, http.StatusNotFound
		}

		pharmacy.Products[idx].Active = false
		return pharmacy, nil, http.StatusNoContent
	})
}
