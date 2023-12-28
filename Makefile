test:
	go test -v ./...

run-server:
	go run cmd/server

run-client:
	go run cmd/client

run-docker:
	docker compose up --force-recreate --build server --build client

lint:
	golangci-lint -v run ./...