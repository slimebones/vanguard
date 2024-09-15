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

	// we have INTEGER PRIMARY KEY instead of SERIAL PRIMARY KEY since sqlite
	// doesn't have auto increment for the latter
	_, err := db.Exec(`
		CREATE TABLE "appuser"(
			"id" INTEGER PRIMARY KEY,
			"hpassword" VARCHAR NOT NULL,
			"username" VARCHAR NOT NULL UNIQUE,
			"firstname" VARCHAR,
			"patronym" VARCHAR,
			"surname" VARCHAR,
			"rt" VARCHAR
		);
	`)
	Unwrap(err)
	// _, err := db.Exec(`TRUNCATE TABLE appuser RESTART IDENTITY`)
	// Unwrap(err)
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
	user := createUser("hello", "1234", "", "", "")

	data := Login{
		Username: user.Username,
		Password: "1234",
	}

	rpc("login", data, server, recorder)
	rt := recorder.Body.String()
	rt = strings.ReplaceAll(rt, `"`, ``)
	token, err := decodeToken(rt, RT_SECRET)
	Unwrap(err)
	Assert(token.Created <= utc())
	Assert(token.UserId == user.Id)

	var inDbRt string
	err = db.QueryRow(
		`SELECT rt FROM appuser WHERE username = 'hello'`,
	).Scan(&inDbRt)
	Unwrap(err)
	Assert(inDbRt == rt)
}

func rpc(
	target string,
	data any,
	server *gin.Engine,
	recorder *httptest.ResponseRecorder,
) *httptest.ResponseRecorder {
	marshal, _ := json.Marshal(data)
	req, _ := http.NewRequest(
		"POST",
		"/rpc/"+target,
		strings.NewReader(string(marshal)),
	)
	server.ServeHTTP(recorder, req)
	Assert(recorder.Code == 200)
	return recorder
}

func TestLogout(t *testing.T) {
	server, recorder := setup()
	user := createUser("hello", "1234", "", "", "")

	data := Login{
		Username: user.Username,
		Password: "1234",
	}
	rpc("login", data, server, recorder)

	rt := recorder.Body.String()
	rt = strings.ReplaceAll(rt, `"`, ``)
	token, err := decodeToken(rt, RT_SECRET)
	Unwrap(err)
	Assert(token.Created <= utc())
	Assert(token.UserId == user.Id)

	var inDbRt string
	err = db.QueryRow(
		`SELECT rt FROM appuser WHERE username = 'hello'`,
	).Scan(&inDbRt)
	Unwrap(err)
	Assert(inDbRt == rt)

	rpc("logout", Logout{Rt: rt}, server, recorder)
	err = db.QueryRow(
		`SELECT ifnull(rt, "") FROM appuser WHERE username = 'hello'`,
	).Scan(&inDbRt)
	Unwrap(err)
	Assert(inDbRt == "")
}
