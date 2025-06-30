# Makefile for simple-sa-token-issuer

.PHONY: build run docker-build docker-run test clean

BINARY_NAME=simple-sa-token-issuer
DOCKER_IMAGE=simple-sa-token-issuer:latest

tidy:
	go mod tidy

build: tidy
	go build -o $(BINARY_NAME) .

run: build
	./$(BINARY_NAME)

lint: tidy
	golangci-lint run --fix

docker-build:
	docker build -t $(DOCKER_IMAGE) .

docker-run: docker-build
	docker run -p 8080:8080 \
		-e AUTH_TOKEN=test-token \
		-e READONLY_SA=readonly \
		-e READONLY_NS=default \
		-e ADMIN_SA=admin \
		-e ADMIN_NS=default \
		$(DOCKER_IMAGE)

compose-up:
	docker-compose up --build

compose-down:
	docker-compose down

test:
	go test -v ./...

clean:
	rm -f $(BINARY_NAME)
	docker rmi $(DOCKER_IMAGE) 2>/dev/null || true