package pharmacy_product

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func orderSeed() *Pharmacy {
	return &Pharmacy{
		Id:   "lekaren-centrum",
		Name: "Lekáreň Centrum",
		Orders: []DepartmentOrder{
			{
				Id:             "ord-1",
				DepartmentName: "Interna A",
				Status:         "created",
				Items: []DepartmentOrderItem{
					{Id: "it-1", ProductName: "Paralen 500 mg", RequestedQty: 5, IssuedQty: 0},
				},
			},
			{
				Id:             "ord-2",
				DepartmentName: "Chirurgia B",
				Status:         "fulfilled",
				Items: []DepartmentOrderItem{
					{Id: "it-2", ProductName: "Betadine", RequestedQty: 2, IssuedQty: 2},
				},
			},
		},
	}
}

func TestGetDepartmentOrders(t *testing.T) {
	db := newFakeDb(orderSeed())
	engine := newTestEngine(db)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/orders/lekaren-centrum/items", nil)
	engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var orders []DepartmentOrder
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &orders))
	assert.Len(t, orders, 2)
}

func TestCreateDepartmentOrder(t *testing.T) {
	db := newFakeDb(orderSeed())
	engine := newTestEngine(db)

	body, _ := json.Marshal(DepartmentOrder{
		DepartmentName: "JIS",
		Items: []DepartmentOrderItem{
			{ProductName: "Vitamín C 1000", RequestedQty: 7, IssuedQty: 0},
		},
	})
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/orders/lekaren-centrum/items", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var created DepartmentOrder
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	assert.Equal(t, "created", created.Status)
	assert.NotEmpty(t, created.Id)
	assert.NotEmpty(t, created.CreatedAt)
}

func TestUpdateDepartmentOrder(t *testing.T) {
	db := newFakeDb(orderSeed())
	engine := newTestEngine(db)

	body, _ := json.Marshal(DepartmentOrder{
		Status: "processing",
		Items: []DepartmentOrderItem{
			{Id: "it-1", ProductName: "Paralen 500 mg", RequestedQty: 5, IssuedQty: 3},
		},
	})
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/orders/lekaren-centrum/items/ord-1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var updated DepartmentOrder
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &updated))
	assert.Equal(t, "processing", updated.Status)
	assert.Equal(t, int32(3), updated.Items[0].IssuedQty)
}

func TestDeleteDepartmentOrder_CancelCreated(t *testing.T) {
	db := newFakeDb(orderSeed())
	engine := newTestEngine(db)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/orders/lekaren-centrum/items/ord-1", nil)
	engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "canceled", db.store["lekaren-centrum"].Orders[0].Status)
}

func TestDeleteDepartmentOrder_ArchiveFulfilled(t *testing.T) {
	db := newFakeDb(orderSeed())
	engine := newTestEngine(db)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/orders/lekaren-centrum/items/ord-2", nil)
	engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "archived", db.store["lekaren-centrum"].Orders[1].Status)
}
