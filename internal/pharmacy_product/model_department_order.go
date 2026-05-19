package pharmacy_product

import (
	"time"
)

type DepartmentOrder struct {
	Id string `json:"id"`

	DepartmentName string `json:"departmentName"`

	Note string `json:"note,omitempty"`

	Status string `json:"status"`

	CreatedAt time.Time `json:"createdAt,omitempty"`

	UpdatedAt time.Time `json:"updatedAt,omitempty"`

	Items []DepartmentOrderItem `json:"items,omitempty"`
}
