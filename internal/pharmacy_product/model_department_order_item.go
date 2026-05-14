package pharmacy_product

type DepartmentOrderItem struct {
	Id string `json:"id"`

	ProductId string `json:"productId,omitempty"`

	ProductName string `json:"productName"`

	RequestedQty int32 `json:"requestedQty"`

	IssuedQty int32 `json:"issuedQty"`
}
