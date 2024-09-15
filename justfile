set shell := ["nu", "-c"]

lint:
    go fmt

test t="":
    if "{{t}}" == "" { go test } else { go test -run {{t}} }

check: lint test

run:
    go run .

migration *t:
	@ dbmate -s "migrations/psql_schema.sql" -d "migrations/psql" -u "postgres://vanguard:vanguard@localhost:9005/vanguard?sslmode=disable" {{t}}
