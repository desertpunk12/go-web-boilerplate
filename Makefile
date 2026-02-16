tg:
	templ generate --watch --proxy="http://localhost:4000" --cmd="go run ./cmd/hrapp-web/main.go"

be:
	go run ./cmd/hrapp-api/main.go

tcss:
	tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css -v --watch --minify


s:
	make -j3 tg tcss be

# Database migrations
# Usage: make migrate-up (apply all pending migrations)
#         make migrate-down (rollback last migration)
#         make migrate-create NAME (create new migration: make migrate-create add_users_table)
#         make migrate-version (show current migration version)

migrate-up:
	migrate -path internal/hr-api/repositories/migrations -database "postgres://$$DATABASE_URL?sslmode=disable" up

migrate-down:
	migrate -path internal/hr-api/repositories/migrations -database "postgres://$$DATABASE_URL?sslmode=disable" down 1

migrate-create:
ifndef NAME
	@echo "Usage: make migrate-create NAME"
	@echo "Example: make migrate-create add_users_table"
else
	migrate create -ext sql -dir internal/hr-api/repositories/migrations -seq -digits 6 $(NAME)
endif

migrate-version:
	migrate -path internal/hr-api/repositories/migrations -database "postgres://$$DATABASE_URL?sslmode=disable" version

migrate-force:
	migrate -path internal/hr-api/repositories/migrations -database "postgres://$$DATABASE_URL?sslmode=disable" force 0

migrate-drop:
	@echo "WARNING: This will drop all tables in the database!"
	migrate -path internal/hr-api/repositories/migrations -database "postgres://$$DATABASE_URL?sslmode=disable" drop -f

