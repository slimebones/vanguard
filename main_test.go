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

func rpc(
	target string,
	data any,
	server *gin.Engine,
	recorder *httptest.ResponseRecorder,
) *httptest.ResponseRecorder {
	marshal, err := json.Marshal(data)
	Unwrap(err)
	req, _ := http.NewRequest(
		"POST",
		"/rpc/"+target,
		strings.NewReader(string(marshal)),
	)
	server.ServeHTTP(recorder, req)
	Assert(recorder.Code == 200)
	return recorder
}

func rpcCompare(
	target string,
	data any,
	server *gin.Engine,
	recorder *httptest.ResponseRecorder,
	compared any,
) *httptest.ResponseRecorder {
	recorder = rpc(target, data, server, recorder)
	comparedJson, err := json.Marshal(compared)
	Unwrap(err)
	Assert(
		recorder.Body.String() == string(comparedJson),
		recorder.Body.String(),
		string(comparedJson),
	)
	return recorder
}

func TestLoginOk(t *testing.T) {
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

func TestLogoutOk(t *testing.T) {
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

func TestGetUsersInOk(t *testing.T) {
	_, _ = setup()
	user1 := createUser("hello", "1234", "", "", "")
	user2 := createUser("world", "1234", "", "", "")
	users, err := getUsers(GetQuery{
		"username": Dict{
			"$in": []string{"hello", "world"},
		},
	})
	Unwrap(err)
	Assert(len(users) == 2)
	Assert(users[0].Username == user1.Username)
	Assert(users[1].Username == user2.Username)
}

func TestGetUsersInAndIdOk(t *testing.T) {
	_, _ = setup()
	user1 := createUser("hello", "1234", "", "", "")
	_ = createUser("world", "1234", "", "", "")
	users, err := getUsers(GetQuery{
		"id": 1,
		"username": Dict{
			"$in": []string{"hello", "world"},
		},
	})
	Unwrap(err)
	Assert(len(users) == 1)
	Assert(users[0].Username == user1.Username)
}

func TestCurrentOk(t *testing.T) {
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

	rpcCompare("current", Current{Rt: rt}, server, recorder, user)
}
