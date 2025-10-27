up:
	docker-compose up -d
	go run ./cmd/frontend/main.go &
	go run ./cmd/server/main.go
down:
	docker-compose down