package pharmacy_product

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type implPharmacyCategoriesAPI struct{}

func NewPharmacyCategoriesApi() PharmacyCategoriesAPI {
	return &implPharmacyCategoriesAPI{}
}

func (o implPharmacyCategoriesAPI) GetCategories(c *gin.Context) {
	updatePharmacyFunc(c, func(_ *gin.Context, pharmacy *Pharmacy) (*Pharmacy, interface{}, int) {
		result := pharmacy.PredefinedCategories
		if result == nil {
			result = []Category{}
		}
		return nil, result, http.StatusOK
	})
}
