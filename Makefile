run/api:
	go run ./cmd/api
db/psql:
	psql ${GREENLIGHT_DB_DSN}
db/migrations/up:
	@echo 'Runing up migrations ...'
	migrate -path ./migrations -database ${GREENLIGHT_DB_DSN} up
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}