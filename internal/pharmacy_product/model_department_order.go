package pharmacy_product

type DepartmentOrder struct {
	Id string `json:"id"`

	DepartmentName string `json:"departmentName"`

	Note string `json:"note,omitempty"`

	Status string `json:"status"`

	CreatedAt string `json:"createdAt,omitempty"`

	UpdatedAt string `json:"updatedAt,omitempty"`

	Items []DepartmentOrderItem `json:"items,omitempty"`
}
