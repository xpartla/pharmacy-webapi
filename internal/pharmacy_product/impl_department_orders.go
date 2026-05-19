package pharmacy_product

import (
	"net/http"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type implDepartmentOrdersAPI struct{}

func NewDepartmentOrdersApi() DepartmentOrdersAPI {
	return &implDepartmentOrdersAPI{}
}

func (o implDepartmentOrdersAPI) GetDepartmentOrders(c *gin.Context) {
	updatePharmacyFunc(c, func(_ *gin.Context, pharmacy *Pharmacy) (*Pharmacy, interface{}, int) {
		result := pharmacy.Orders
		if result == nil {
			result = []DepartmentOrder{}
		}
		return nil, result, http.StatusOK
	})
}

func (o implDepartmentOrdersAPI) GetDepartmentOrder(c *gin.Context) {
	updatePharmacyFunc(c, func(_ *gin.Context, pharmacy *Pharmacy) (*Pharmacy, interface{}, int) {
		orderId := c.Param("orderId")
		if orderId == "" {
			return nil, gin.H{"status": http.StatusBadRequest, "message": "Order ID is required"}, http.StatusBadRequest
		}

		idx := slices.IndexFunc(pharmacy.Orders, func(order DepartmentOrder) bool { return order.Id == orderId })
		if idx < 0 {
			return nil, gin.H{"status": http.StatusNotFound, "message": "Department order not found"}, http.StatusNotFound
		}
		return nil, pharmacy.Orders[idx], http.StatusOK
	})
}

func (o implDepartmentOrdersAPI) CreateDepartmentOrder(c *gin.Context) {
	updatePharmacyFunc(c, func(_ *gin.Context, pharmacy *Pharmacy) (*Pharmacy, interface{}, int) {
		var order DepartmentOrder
		if err := c.ShouldBindJSON(&order); err != nil {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Invalid request body",
				"error":   err.Error(),
			}, http.StatusBadRequest
		}

		if order.DepartmentName == "" {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Department name is required",
			}, http.StatusBadRequest
		}

		if len(order.Items) == 0 {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Order must contain at least one item",
			}, http.StatusBadRequest
		}

		for _, item := range order.Items {
			if item.ProductName == "" {
				return nil, gin.H{
					"status":  http.StatusBadRequest,
					"message": "Each order item must have productName",
				}, http.StatusBadRequest
			}
			if item.RequestedQty <= 0 {
				return nil, gin.H{
					"status":  http.StatusBadRequest,
					"message": "requestedQty must be greater than zero",
				}, http.StatusBadRequest
			}
			if item.IssuedQty < 0 {
				return nil, gin.H{
					"status":  http.StatusBadRequest,
					"message": "issuedQty must be zero or greater",
				}, http.StatusBadRequest
			}
		}

		if order.Id == "" || order.Id == "@new" {
			order.Id = uuid.NewString()
		}
		conflict := slices.IndexFunc(pharmacy.Orders, func(existing DepartmentOrder) bool { return existing.Id == order.Id })
		if conflict >= 0 {
			return nil, gin.H{
				"status":  http.StatusConflict,
				"message": "Department order with this id already exists",
			}, http.StatusConflict
		}

		now := time.Now().UTC()
		order.Status = "created"
		order.CreatedAt = now
		order.UpdatedAt = now
		for i := range order.Items {
			if order.Items[i].Id == "" {
				order.Items[i].Id = uuid.NewString()
			}
		}

		pharmacy.Orders = append(pharmacy.Orders, order)
		return pharmacy, order, http.StatusOK
	})
}

func (o implDepartmentOrdersAPI) UpdateDepartmentOrder(c *gin.Context) {
	updatePharmacyFunc(c, func(_ *gin.Context, pharmacy *Pharmacy) (*Pharmacy, interface{}, int) {
		orderId := c.Param("orderId")
		if orderId == "" {
			return nil, gin.H{"status": http.StatusBadRequest, "message": "Order ID is required"}, http.StatusBadRequest
		}

		var incoming DepartmentOrder
		if err := c.ShouldBindJSON(&incoming); err != nil {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Invalid request body",
				"error":   err.Error(),
			}, http.StatusBadRequest
		}

		idx := slices.IndexFunc(pharmacy.Orders, func(order DepartmentOrder) bool { return order.Id == orderId })
		if idx < 0 {
			return nil, gin.H{"status": http.StatusNotFound, "message": "Department order not found"}, http.StatusNotFound
		}

		existing := pharmacy.Orders[idx]
		if incoming.DepartmentName != "" {
			existing.DepartmentName = incoming.DepartmentName
		}
		existing.Note = incoming.Note

		if incoming.Status != "" {
			if !isValidDepartmentOrderStatus(incoming.Status) {
				return nil, gin.H{"status": http.StatusBadRequest, "message": "Invalid status"}, http.StatusBadRequest
			}
			existing.Status = incoming.Status
		}

		if incoming.Items != nil {
			if len(incoming.Items) == 0 {
				return nil, gin.H{"status": http.StatusBadRequest, "message": "Order must contain at least one item"}, http.StatusBadRequest
			}
			for i := range incoming.Items {
				if incoming.Items[i].ProductName == "" {
					return nil, gin.H{"status": http.StatusBadRequest, "message": "Each order item must have productName"}, http.StatusBadRequest
				}
				if incoming.Items[i].RequestedQty <= 0 {
					return nil, gin.H{"status": http.StatusBadRequest, "message": "requestedQty must be greater than zero"}, http.StatusBadRequest
				}
				if incoming.Items[i].IssuedQty < 0 {
					return nil, gin.H{"status": http.StatusBadRequest, "message": "issuedQty must be zero or greater"}, http.StatusBadRequest
				}
				if incoming.Items[i].Id == "" {
					incoming.Items[i].Id = uuid.NewString()
				}
			}
			existing.Items = incoming.Items
		}

		existing.UpdatedAt = time.Now().UTC()
		pharmacy.Orders[idx] = existing
		return pharmacy, existing, http.StatusOK
	})
}

func (o implDepartmentOrdersAPI) DeleteDepartmentOrder(c *gin.Context) {
	updatePharmacyFunc(c, func(_ *gin.Context, pharmacy *Pharmacy) (*Pharmacy, interface{}, int) {
		orderId := c.Param("orderId")
		if orderId == "" {
			return nil, gin.H{"status": http.StatusBadRequest, "message": "Order ID is required"}, http.StatusBadRequest
		}

		idx := slices.IndexFunc(pharmacy.Orders, func(order DepartmentOrder) bool { return order.Id == orderId })
		if idx < 0 {
			return nil, gin.H{"status": http.StatusNotFound, "message": "Department order not found"}, http.StatusNotFound
		}

		switch pharmacy.Orders[idx].Status {
		case "created":
			pharmacy.Orders[idx].Status = "canceled"
		case "fulfilled":
			pharmacy.Orders[idx].Status = "archived"
		default:
			return nil, gin.H{
				"status":  http.StatusConflict,
				"message": "Order can be canceled only in created state or archived only in fulfilled state",
			}, http.StatusConflict
		}
		pharmacy.Orders[idx].UpdatedAt = time.Now().UTC()
		return pharmacy, nil, http.StatusNoContent
	})
}

func isValidDepartmentOrderStatus(status string) bool {
	switch status {
	case "created", "processing", "fulfilled", "canceled", "archived":
		return true
	default:
		return false
	}
}
