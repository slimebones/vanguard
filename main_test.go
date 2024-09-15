package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func migrateDb() {
	// file, err := os.Open("migrations/psql_schema.sql")
	// defer func() {
	// 	err = file.Close()
	// 	Unwrap(err)
	// }()
	// Unwrap(err)
	// b, err := io.ReadAll(file)
	// Unwrap(err)

	// content := string(b)
	// _, err = db.Query(content)
	// Unwrap(err)

	_, err := db.Query(`
		CREATE TABLE "appuser"(
			"id" SERIAL PRIMARY KEY,
			"hpassword" VARCHAR NOT NULL,
			"username" VARCHAR NOT NULL UNIQUE,
			"firstname" VARCHAR,
			"patronym" VARCHAR,
			"surname" VARCHAR,
			"rt" VARCHAR
		);
	`)
	Unwrap(err)
}

func setup() (*gin.Engine, *httptest.ResponseRecorder) {
	server := newServer(
		NewServerArgs{dbDriver: "sqlite", dbUrl: ":memory:"},
	)
	recorder := httptest.NewRecorder()
	migrateDb()
	return server, recorder
}

func TestLogin(t *testing.T) {
	server, recorder := setup()
	createUser("hello", "1234", "", "", "")
	Assert(false)

	data := Login{
		Username: "hello",
		Password: "1234",
	}
	marshal, _ := json.Marshal(data)

	req, _ := http.NewRequest(
		"POST",
		"/rpc/login",
		strings.NewReader(string(marshal)),
	)
	server.ServeHTTP(recorder, req)
	Assert(recorder.Code == 200)
	Assert(len(recorder.Body.String()) > 0)
	Assert(false)
}
