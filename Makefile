# .PHONY: run

# run:
# 	go run cmd/students-api/main.go -config config/local.yaml


# Default configuration path if none is supplied
CONFIG ?= config/local.yaml

.PHONY: run
run:
	@go run cmd/students-api/main.go -config $(CONFIG)