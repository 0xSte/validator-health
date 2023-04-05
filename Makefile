default: gen-oapi

gen-oapi:
	mkdir -p pkg/beacon
	swagger generate client --spec api/openapi.json
	go mod tidy
