package pharmacy_product

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/xpartla/pharmacy-webapi/internal/db_service"
)

type implPharmaciesAPI struct{}

func NewPharmaciesApi() PharmaciesAPI {
	return &implPharmaciesAPI{}
}

func (o implPharmaciesAPI) CreatePharmacy(c *gin.Context) {
	value, exists := c.Get("db_service")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Internal Server Error",
			"message": "db_service not found",
			"error":   "db_service not found",
		})
		return
	}

	db, ok := value.(db_service.DbService[Pharmacy])
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Internal Server Error",
			"message": "db context is not of required type",
			"error":   "cannot cast db context to db_service.DbService",
		})
		return
	}

	pharmacy := Pharmacy{}
	if err := c.BindJSON(&pharmacy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	if pharmacy.Id == "" {
		pharmacy.Id = uuid.New().String()
	}

	err := db.CreateDocument(c, pharmacy.Id, &pharmacy)
	switch err {
	case nil:
		c.JSON(http.StatusOK, pharmacy)
	case db_service.ErrConflict:
		c.JSON(http.StatusConflict, gin.H{
			"status":  "Conflict",
			"message": "Pharmacy already exists",
			"error":   err.Error(),
		})
	default:
		c.JSON(http.StatusBadGateway, gin.H{
			"status":  "Bad Gateway",
			"message": "Failed to create pharmacy in database",
			"error":   err.Error(),
		})
	}
}

func (o implPharmaciesAPI) DeletePharmacy(c *gin.Context) {
	value, exists := c.Get("db_service")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Internal Server Error",
			"message": "db_service not found",
			"error":   "db_service not found",
		})
		return
	}

	db, ok := value.(db_service.DbService[Pharmacy])
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Internal Server Error",
			"message": "db_service context is not of type db_service.DbService",
			"error":   "cannot cast db_service context to db_service.DbService",
		})
		return
	}

	pharmacyId := c.Param("pharmacyId")
	err := db.DeleteDocument(c, pharmacyId)
	switch err {
	case nil:
		c.AbortWithStatus(http.StatusNoContent)
	case db_service.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "Not Found",
			"message": "Pharmacy not found",
			"error":   err.Error(),
		})
	default:
		c.JSON(http.StatusBadGateway, gin.H{
			"status":  "Bad Gateway",
			"message": "Failed to delete pharmacy from database",
			"error":   err.Error(),
		})
	}
}
