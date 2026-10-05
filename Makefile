SWAG ?= $(shell go env GOPATH)/bin/swag

.PHONY: swagger swagger-install

# Generate OpenAPI/Swagger spec (json + yaml) into docs/swagger.
# docs.go is skipped on purpose: it imports swaggo/swag, which is not vendored.
swagger:
	$(SWAG) init -g cmd/servers/main/main.go -o docs/swagger --outputTypes json,yaml --exclude vendor

swagger-install:
	go install github.com/swaggo/swag/cmd/swag@latest
