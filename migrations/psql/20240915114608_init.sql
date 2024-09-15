-- migrate:up
CREATE TABLE "appuser"(
	"id" SERIAL PRIMARY KEY,
	"hpassword" VARCHAR NOT NULL,
	"username" VARCHAR NOT NULL UNIQUE,
	"firstname" VARCHAR,
	"patronym" VARCHAR,
	"surname" VARCHAR,
	"rt" VARCHAR
);

-- migrate:down

