set shell := ["nu", "-c"]
set dotenv-load

lint:
    go fmt

test t="":
    if "{{t}}" == "" { go test } else { go test -run {{t}} }

check: lint test

run:
    go run .

docker_build:
    docker-compose up -d --build --remove-orphans

migration *t:
	@ dbmate -s "migrations/psql_schema.sql" -d "migrations/psql" {{t}}
