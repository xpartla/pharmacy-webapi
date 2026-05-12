# pharmacy-webapi

Go + Gin REST API backing the `pharmacy-ufe` micro-frontend. Persists `Pharmacy` documents in MongoDB; each document embeds its own product list and predefined category list. Soft delete on products (DELETE flips `active=false`).

## Endpoints

```
POST   /api/pharmacy
DELETE /api/pharmacy/:pharmacyId
GET    /api/products/:pharmacyId/items          # only active products
POST   /api/products/:pharmacyId/items
GET    /api/products/:pharmacyId/items/:productId
PUT    /api/products/:pharmacyId/items/:productId
DELETE /api/products/:pharmacyId/items/:productId   # soft delete (active=false)
GET    /api/products/:pharmacyId/categories
GET    /openapi
```

## Develop

```bash
# Regenerate Gin interface stubs (only when api/pharmacy-product.openapi.yaml changes)
docker run --rm --user "$(id -u):$(id -g)" -v "$PWD":/local \
    openapitools/openapi-generator-cli:latest generate -c /local/scripts/generator-cfg.yaml
# (then delete the generator's main.go, Dockerfile, go.mod, api/openapi.yaml, internal/pharmacy_product/README.md)

go mod tidy
go test ./...

# Local Mongo + Mongo Express
docker compose -f deployments/docker-compose/compose.yaml up -d
PHARMACY_API_MONGODB_USERNAME=root PHARMACY_API_MONGODB_PASSWORD=neUhaDnes \
    go run ./cmd/pharmacy-api-service
```

## Environment variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `PHARMACY_API_PORT` | `8080` | HTTP listen port |
| `PHARMACY_API_ENVIRONMENT` | `production` | Non-production → Gin debug mode |
| `PHARMACY_API_MONGODB_HOST` | `localhost` | Mongo host |
| `PHARMACY_API_MONGODB_PORT` | `27017` | Mongo port |
| `PHARMACY_API_MONGODB_USERNAME` | (empty) | Mongo auth |
| `PHARMACY_API_MONGODB_PASSWORD` | (empty) | Mongo auth |
| `PHARMACY_API_MONGODB_DATABASE` | `xpartla-pharmacy` | Database |
| `PHARMACY_API_MONGODB_COLLECTION` | `pharmacy` | Collection |
| `PHARMACY_API_MONGODB_TIMEOUT_SECONDS` | `10` | Operation timeout |
| `LOG_LEVEL` | `info` | zerolog level |

## Docker build

```bash
docker build -t xpartla/pharmacy-webapi:local -f build/docker/Dockerfile .
```
