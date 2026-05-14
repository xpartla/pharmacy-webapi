package pharmacy_product

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/xpartla/pharmacy-webapi/internal/db_service"
)

// fakeDb is a minimal in-memory db_service.DbService used for handler tests.
type fakeDb struct {
	store map[string]*Pharmacy
}

func newFakeDb(seed *Pharmacy) *fakeDb {
	return &fakeDb{store: map[string]*Pharmacy{seed.Id: seed}}
}

func (f *fakeDb) CreateDocument(_ context.Context, id string, doc *Pharmacy) error {
	if _, exists := f.store[id]; exists {
		return db_service.ErrConflict
	}
	f.store[id] = doc
	return nil
}
func (f *fakeDb) FindDocument(_ context.Context, id string) (*Pharmacy, error) {
	if doc, ok := f.store[id]; ok {
		return doc, nil
	}
	return nil, db_service.ErrNotFound
}
func (f *fakeDb) UpdateDocument(_ context.Context, id string, doc *Pharmacy) error {
	if _, ok := f.store[id]; !ok {
		return db_service.ErrNotFound
	}
	f.store[id] = doc
	return nil
}
func (f *fakeDb) DeleteDocument(_ context.Context, id string) error {
	if _, ok := f.store[id]; !ok {
		return db_service.ErrNotFound
	}
	delete(f.store, id)
	return nil
}
func (f *fakeDb) Disconnect(_ context.Context) error { return nil }

func newTestEngine(db db_service.DbService[Pharmacy]) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) { c.Set("db_service", db); c.Next() })
	handlers := ApiHandleFunctions{
		PharmaciesAPI:         NewPharmaciesApi(),
		PharmacyCategoriesAPI: NewPharmacyCategoriesApi(),
		PharmacyProductsAPI:   NewPharmacyProductsApi(),
		DepartmentOrdersAPI:   NewDepartmentOrdersApi(),
	}
	NewRouterWithGinEngine(engine, handlers)
	return engine
}

func sampleSeed() *Pharmacy {
	return &Pharmacy{
		Id:   "lekaren-centrum",
		Name: "Lekáreň Centrum",
		PredefinedCategories: []Category{
			{Code: "analgesics", Value: "Analgetiká"},
			{Code: "vitamins", Value: "Vitamíny"},
		},
		Products: []Product{
			{Id: "p1", Name: "Paralen", Stock: 10, Active: true, Category: Category{Code: "analgesics", Value: "Analgetiká"}},
			{Id: "p2", Name: "Old Drug", Stock: 0, Active: false, Category: Category{Code: "analgesics", Value: "Analgetiká"}},
		},
	}
}

func TestGetProducts_HidesInactive(t *testing.T) {
	db := newFakeDb(sampleSeed())
	engine := newTestEngine(db)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/products/lekaren-centrum/items", nil)
	engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var products []Product
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &products))
	assert.Len(t, products, 1)
	assert.Equal(t, "p1", products[0].Id)
}

func TestGetCategories(t *testing.T) {
	db := newFakeDb(sampleSeed())
	engine := newTestEngine(db)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/products/lekaren-centrum/categories", nil)
	engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var cats []Category
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &cats))
	assert.Len(t, cats, 2)
}

func TestCreateProduct_DefaultsActive(t *testing.T) {
	db := newFakeDb(sampleSeed())
	engine := newTestEngine(db)

	body, _ := json.Marshal(Product{Name: "Vitamín C", Stock: 5})
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/products/lekaren-centrum/items", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var created Product
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	assert.True(t, created.Active)
	assert.NotEmpty(t, created.Id)
	assert.Equal(t, "Vitamín C", created.Name)
	assert.Equal(t, int32(5), created.Stock)
}

func TestUpdateProduct_ChangesStock(t *testing.T) {
	db := newFakeDb(sampleSeed())
	engine := newTestEngine(db)

	body, _ := json.Marshal(Product{Stock: 99, Active: true})
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/products/lekaren-centrum/items/p1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var updated Product
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &updated))
	assert.Equal(t, int32(99), updated.Stock)
}

func TestDeleteProduct_SoftDeletes(t *testing.T) {
	db := newFakeDb(sampleSeed())
	engine := newTestEngine(db)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/products/lekaren-centrum/items/p1", nil)
	engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)

	// Listing should now hide p1 because we soft-deleted it.
	listRec := httptest.NewRecorder()
	listReq, _ := http.NewRequest(http.MethodGet, "/api/products/lekaren-centrum/items", nil)
	engine.ServeHTTP(listRec, listReq)
	var visible []Product
	assert.NoError(t, json.Unmarshal(listRec.Body.Bytes(), &visible))
	assert.Empty(t, visible)

	// But the document itself still exists in the pharmacy with active=false.
	found := db.store["lekaren-centrum"]
	assert.Len(t, found.Products, 2)
	for _, p := range found.Products {
		if p.Id == "p1" {
			assert.False(t, p.Active, "p1 must be soft-deleted, not removed")
		}
	}
}

func TestGetProduct_NotFound(t *testing.T) {
	db := newFakeDb(sampleSeed())
	engine := newTestEngine(db)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/products/lekaren-centrum/items/does-not-exist", nil)
	engine.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetProducts_IncludeInactive(t *testing.T) {
	db := newFakeDb(sampleSeed())
	engine := newTestEngine(db)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/products/lekaren-centrum/items?include=inactive", nil)
	engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var products []Product
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &products))
	assert.Len(t, products, 1)
	assert.Equal(t, "p2", products[0].Id)
	assert.False(t, products[0].Active)
}

func TestGetProducts_IncludeAll(t *testing.T) {
	db := newFakeDb(sampleSeed())
	engine := newTestEngine(db)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/products/lekaren-centrum/items?include=all", nil)
	engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var products []Product
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &products))
	assert.Len(t, products, 2)
}

func TestGetProducts_IncludeInvalid(t *testing.T) {
	db := newFakeDb(sampleSeed())
	engine := newTestEngine(db)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/products/lekaren-centrum/items?include=garbage", nil)
	engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateProduct_Reactivate(t *testing.T) {
	// p2 starts inactive in the seed — PUT with active=true should restore it.
	db := newFakeDb(sampleSeed())
	engine := newTestEngine(db)

	body, _ := json.Marshal(Product{Active: true, Stock: 1})
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/products/lekaren-centrum/items/p2", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var updated Product
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &updated))
	assert.True(t, updated.Active, "PUT should be able to flip active back to true")
}

func TestCreatePharmacy_Conflict(t *testing.T) {
	// Seed already contains lekaren-centrum; POSTing the same id must 409.
	db := newFakeDb(sampleSeed())
	engine := newTestEngine(db)

	body, _ := json.Marshal(Pharmacy{Id: "lekaren-centrum", Name: "Duplicate"})
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/pharmacy", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestDeletePharmacy_NotFound(t *testing.T) {
	db := newFakeDb(sampleSeed())
	engine := newTestEngine(db)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/pharmacy/no-such-pharmacy", nil)
	engine.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}
