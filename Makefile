.PHONY: run
run:
	@echo "Building the application..."
	@go build -gcflags='all=-N -l' -o aibuddy main.go
	@./aibuddy

.PHONY: test_coverage
test_coverage:
	go test -cover -coverprofile=coverage.out -gcflags='all=-N -l' -coverpkg ./... -timeout=5m ./... -cpu=1
	@echo "Coverage report:"
	@go tool cover -html=coverage.out
	@echo "Coverage report generated successfully"

.PHONY: lint
lint:
	@# scoop install main/golangci-lint
	golangci-lint run ./...

.PHONY: gen
gen: 
	@go run cmd/generate/generate.go