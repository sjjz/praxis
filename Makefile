APP_DIR=app
DB_URL=postgres://postgres:postgres@localhost:5432/praxis?sslmode=disable
GO_IMAGE=golang:1.23.6

.PHONY: run test migrate sqlc docker-up docker-down

run:
	docker run --rm -it \
		-v $(PWD)/$(APP_DIR):/src \
		-w /src \
		-p 8080:8080 \
		-e HTTP_ADDR=:8080 \
		-e DATABASE_URL=$(DB_URL) \
		-e DEV_USER_ID=00000000-0000-0000-0000-000000000001 \
		$(GO_IMAGE) bash -lc "export PATH=/usr/local/go/bin:$$PATH && go mod tidy && go run ."

test:
	docker run --rm -t \
		-v $(PWD)/$(APP_DIR):/src \
		-w /src \
		$(GO_IMAGE) bash -lc "export PATH=/usr/local/go/bin:$$PATH && go mod tidy && go test ./..."

migrate:
	@for f in $(APP_DIR)/db/migrations/*.sql; do \
		echo "Applying $$f"; \
		cat $$f | docker compose exec -T db psql -U postgres -d praxis; \
	done

sqlc:
	docker run --rm \
		-v $(PWD)/$(APP_DIR):/src \
		-w /src \
		sqlc/sqlc:1.28.0 generate -f db/sqlc.yaml

docker-up:
	docker compose up -d db

docker-down:
	docker compose down -v
