package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// Migrates db for testing purposes.
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

	// _, err := db.Exec(`
	// 	CREATE TABLE "appuser"(
	// 		"id" SERIAL PRIMARY KEY,
	// 		"hpassword" VARCHAR NOT NULL,
	// 		"username" VARCHAR NOT NULL UNIQUE,
	// 		"firstname" VARCHAR,
	// 		"patronym" VARCHAR,
	// 		"surname" VARCHAR,
	// 		"rt" VARCHAR
	// 	);
	// `)
	// Unwrap(err)
	_, err := db.Exec(`TRUNCATE TABLE appuser RESTART IDENTITY`)
	Unwrap(err)
}

func setup() (*gin.Engine, *httptest.ResponseRecorder) {
	server := newServer(
		NewServerArgs{},
	)
	recorder := httptest.NewRecorder()
	migrateDb()
	return server, recorder
}

func TestLogin(t *testing.T) {
	server, recorder := setup()
	user := createUser("hello", "1234", "", "", "")

	data := Login{
		Username: user.Username,
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
	rt := recorder.Body.String()
	rt = strings.ReplaceAll(rt, `"`, ``)
	token, err := decodeToken(rt, RT_SECRET)
	Unwrap(err)
	Assert(token.Created <= utc())
	Assert(token.UserId == user.Id)
}
