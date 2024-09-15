lint:
    go fmt

test:
    go test

check: lint test

run:
    go run .

migration *t:
	@ dbmate -s "migrations/psql_schema.sql" -d "migrations/psql" -u "postgres://vanguard:vanguard@localhost:9005/vanguard?sslmode=disable" {{t}}
