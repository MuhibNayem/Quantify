.PHONY: dev dev-down dev-logs migrate seed test test-integration lint build docker-build k8s-deploy tf-plan tf-apply clean

dev:
	docker compose up -d --build

dev-down:
	docker compose down

dev-logs:
	docker compose logs -f

migrate:
	docker compose exec backend go run ./cmd/migrator -dsn "postgres://$${POSTGRES_USER:-user}:$${POSTGRES_PASSWORD:-password}@db:5432/$${POSTGRES_DB:-inventory_db}?sslmode=disable" -cmd up

seed:
	./scripts/seed-db.sh

test:
	cd backend && go test ./... -count=1
	cd ai-service && python -m pytest

test-integration:
	cd backend && go test ./... -tags=integration -count=1

lint:
	cd backend && golangci-lint run ./...
	cd ai-service && ruff check .

build:
	cd backend && CGO_ENABLED=0 go build -o ../bin/backend ./cmd/server/main.go
	cd backend && CGO_ENABLED=0 go build -o ../bin/migrator ./cmd/migrator/main.go

docker-build:
	docker compose build

k8s-deploy:
	kubectl apply -f k8s/

tf-plan:
	cd terraform && terraform plan

tf-apply:
	cd terraform && terraform apply

clean:
	rm -rf bin/
	docker compose down -v --remove-orphans
	docker system prune -f
