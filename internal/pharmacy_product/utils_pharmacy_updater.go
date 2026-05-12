package pharmacy_product

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/xpartla/pharmacy-webapi/internal/db_service"
)

type pharmacyUpdater = func(
	ctx *gin.Context,
	pharmacy *Pharmacy,
) (updatedPharmacy *Pharmacy, responseContent interface{}, status int)

// updatePharmacyFunc loads the pharmacy referenced by the :pharmacyId path param,
// runs `updater` against it, and saves the result back to MongoDB. If the updater
// returns a nil *Pharmacy the DB write is skipped (read-only operations).
func updatePharmacyFunc(ctx *gin.Context, updater pharmacyUpdater) {
	value, exists := ctx.Get("db_service")
	if !exists {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Internal Server Error",
			"message": "db_service not found",
			"error":   "db_service not found",
		})
		return
	}

	db, ok := value.(db_service.DbService[Pharmacy])
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Internal Server Error",
			"message": "db_service context is not of type db_service.DbService",
			"error":   "cannot cast db_service context to db_service.DbService",
		})
		return
	}

	pharmacyId := ctx.Param("pharmacyId")
	pharmacy, err := db.FindDocument(ctx, pharmacyId)

	switch err {
	case nil:
	case db_service.ErrNotFound:
		ctx.JSON(http.StatusNotFound, gin.H{
			"status":  "Not Found",
			"message": "Pharmacy not found",
			"error":   err.Error(),
		})
		return
	default:
		ctx.JSON(http.StatusBadGateway, gin.H{
			"status":  "Bad Gateway",
			"message": "Failed to load pharmacy from database",
			"error":   err.Error(),
		})
		return
	}

	updatedPharmacy, responseObject, status := updater(ctx, pharmacy)

	if updatedPharmacy != nil {
		err = db.UpdateDocument(ctx, pharmacyId, updatedPharmacy)
	} else {
		err = nil
	}

	switch err {
	case nil:
		if responseObject != nil {
			ctx.JSON(status, responseObject)
		} else {
			ctx.AbortWithStatus(status)
		}
	case db_service.ErrNotFound:
		ctx.JSON(http.StatusNotFound, gin.H{
			"status":  "Not Found",
			"message": "Pharmacy was deleted while processing the request",
			"error":   err.Error(),
		})
	default:
		ctx.JSON(http.StatusBadGateway, gin.H{
			"status":  "Bad Gateway",
			"message": "Failed to update pharmacy in database",
			"error":   err.Error(),
		})
	}
}
