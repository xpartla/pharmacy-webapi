# Stage 1: regenerate the OpenAPI server interfaces so the build never depends on
# whatever was committed to internal/pharmacy_product/api_*.go and routers.go.
FROM openapitools/openapi-generator-cli:latest AS api

WORKDIR /local

COPY api api
COPY scripts scripts
COPY .openapi-generator-ignore .openapi-generator-ignore

RUN docker-entrypoint.sh generate -c /local/scripts/generator-cfg.yaml

ENTRYPOINT ["bash"]

############################################

FROM golang:latest AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY internal/ internal/
COPY cmd/ cmd/
COPY --from=api /local/ ./

RUN go test ./...

# Build a statically linked binary so the scratch image runs without libc.
RUN CGO_ENABLED=0 GOOS=linux \
      go build \
      -ldflags="-w -s" \
      -installsuffix 'static' \
      -o ./pharmacy-webapi-srv ./cmd/pharmacy-api-service

############################################
FROM scratch

LABEL org.opencontainers.image.authors="Adam Partl"
LABEL org.opencontainers.image.title="Pharmacy Product WebAPI Service"
LABEL org.opencontainers.image.description="WEBAPI for managing pharmacy product inventory"

ENV PHARMACY_API_ENVIRONMENT=production
ENV PHARMACY_API_PORT=8080
ENV PHARMACY_API_MONGODB_HOST=mongo
ENV PHARMACY_API_MONGODB_PORT=27017
ENV PHARMACY_API_MONGODB_DATABASE=xpartla-pharmacy
ENV PHARMACY_API_MONGODB_COLLECTION=pharmacy
ENV PHARMACY_API_MONGODB_USERNAME=root
ENV PHARMACY_API_MONGODB_PASSWORD=
ENV PHARMACY_API_MONGODB_TIMEOUT_SECONDS=5
ENV LOG_LEVEL=info

COPY --from=build /app/pharmacy-webapi-srv ./

EXPOSE 8080
ENTRYPOINT ["./pharmacy-webapi-srv"]
