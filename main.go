package main

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	_ "github.com/lib/pq"

	_ "github.com/glebarez/go-sqlite"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

var dbDriver string

const DEFAULT_DB_DRIVER = "postgres"
const DEFAULT_DB_URL = "postgres://vanguard:vanguard@localhost:9005/vanguard?sslmode=disable"
const RT_SECRET = "weloveauth"
const AT_SECRET = "helloworld"

type Id = int64
type Time = int64
type Dict = map[string]any
type Query = Dict
type GetQuery = Query

func utc() Time {
	return time.Now().Unix()
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func encodeToken(secret string, userId Id) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userId,
		"created": utc(),
	})
	return token.SignedString([]byte(secret))
}

func Err(msg string) error {
	return errors.New(msg)
}

type Token struct {
	UserId  Id   `json:"user_id"`
	Created Time `json:"created"`
	jwt.StandardClaims
}

func decodeToken(token string, secret string) (*Token, error) {
	jwtToken, err := jwt.ParseWithClaims(
		token,
		&Token{},
		func(t *jwt.Token) (any, error) { return []byte(secret), nil },
	)
	if err != nil {
		return nil, err
	}

	claims, ok := jwtToken.Claims.(*Token)
	if !ok {
		return nil, Err("cannot retrieve token claims")
	}
	return claims, nil
}

type User struct {
	Id        Id     `json:"id"`
	Username  string `json:"username"`
	Firstname string `json:"firstname"`
	Patronym  string `json:"patronym"`
	Surname   string `json:"surname"`
	Rt        string `json:"rt"`
}

func createUser(
	username string,
	password string,
	firstname string,
	patronym string,
	surname string,
) User {
	hpassword, err := hashPassword(password)
	Unwrap(err)

	_, err = db.Exec(
		`
			INSERT INTO appuser (
				username, hpassword, firstname, patronym, surname
			) VALUES
				($1, $2, $3, $4, $5)
		`,
		username,
		hpassword,
		firstname,
		patronym,
		surname,
	)
	Unwrap(err)

	var id Id
	err = db.QueryRow(
		`
			SELECT id FROM appuser WHERE username = $1
		`,
		username,
	).Scan(&id)
	Unwrap(err)

	return User{
		Id:        id,
		Username:  username,
		Firstname: firstname,
		Patronym:  patronym,
		Surname:   surname,
		Rt:        "",
	}
}

type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Unwrap(err error) {
	if err != nil {
		panic(err)
	}
}

func Print(obj ...any) {
	fmt.Println(obj...)
}

func RpcLogin(c *gin.Context) {
	var data Login
	err := c.BindJSON(&data)
	Unwrap(err)

	var id Id
	var hpassword string
	err = db.QueryRow(
		`SELECT id, hpassword FROM appuser WHERE username = $1`,
		data.Username,
	).Scan(&id, &hpassword)
	if err != nil {
		panic("invalid username")
	}

	if !checkPasswordHash(data.Password, hpassword) {
		panic("invalid password")
	}

	rt, err := encodeToken(RT_SECRET, id)
	Unwrap(err)

	_, err = db.Exec(
		`
			UPDATE appuser
			SET rt = $1
			WHERE username = $2
		`,
		rt,
		data.Username,
	)
	Unwrap(err)

	c.JSON(200, rt)
}

type Logout struct {
	Rt string `json:"rt"`
}

type Current struct {
	Rt string `json:"rt"`
}

func RpcLogout(c *gin.Context) {
	var data Logout
	err := c.BindJSON(&data)
	Unwrap(err)

	_, err = db.Exec(`UPDATE appuser SET rt = NULL WHERE rt = $1`, data.Rt)
	Unwrap(err)
	c.JSON(200, gin.H{})
}

func RpcCurrent(c *gin.Context) {
	var data Current
	err := c.BindJSON(&data)
	Unwrap(err)

	users, err := getUsers(GetQuery{
		"rt": data.Rt,
	})
	Unwrap(err)
	if len(users) == 0 {
		panic("No users with such refresh token.")
	}
	user := users[0]

	c.JSON(200, user)
}

type Access struct {
	Rt string `json:"rt"`
}

func RpcAccess(c *gin.Context) {
	var data Access
	err := c.BindJSON(&data)
	Unwrap(err)

	users, err := getUsers(GetQuery{
		"rt": data.Rt,
	})
	Unwrap(err)
	if len(users) == 0 {
		panic("No users with such refresh token.")
	}
	user := users[0]
	at, err := encodeToken(AT_SECRET, user.Id)
	Unwrap(err)

	c.JSON(200, at)
}

func RpcReg(c *gin.Context) {
	c.JSON(404, gin.H{})
}

func RpcDereg(c *gin.Context) {
	c.JSON(404, gin.H{})
}

// ref: https://stackoverflow.com/a/71624929/14748231
func Map[T, U any](ts []T, f func(T) U) []U {
	us := make([]U, len(ts))
	for i := range ts {
		us[i] = f(ts[i])
	}
	return us
}

// Ref: https://stackoverflow.com/a/73029665/14748231
func unpackArr(s any) []any {
	v := reflect.ValueOf(s)
	r := make([]any, v.Len())
	for i := 0; i < v.Len(); i++ {
		r[i] = v.Index(i).Interface()
	}
	return r
}

func getUsers(gq GetQuery) ([]User, error) {
	// Extreme levels of sql injection danger are in the air. But we're ok for
	// now.
	q := `SELECT id, username, firstname, patronym, surname, ifnull(rt, "") FROM appuser WHERE `
	var qArgs []any
	for k, v := range gq {
		if !strings.HasSuffix(q, "WHERE ") {
			q += " AND "
		}
		if strings.HasPrefix(k, "$") {
			return nil, Err("Cannot have top-level operator.")
		}
		if v, ok := v.(GetQuery); ok {
			for k2, v2 := range v {
				if k2 != "$in" {
					return nil, Err(
						"Only $in is supported as second-level operator.",
					)
				}
				arr := unpackArr(v2)
				strArr := Map(
					arr,
					func(x any) string {
						xStr, ok := x.(string)
						if !ok {
							panic("Only strings are supported for $in.")
						}
						return "'" + xStr + "'"
					},
				)
				joined := strings.Join(
					strArr,
					", ",
				)
				q += k + " IN (" + joined + ")"
			}
			continue
		}
		qArgs = append(qArgs, v)
		argNumStr := strconv.Itoa(len(qArgs))
		q += k + ` = $` + argNumStr
	}
	q += ";"
	rows, err := db.Query(q, qArgs...)
	Unwrap(err)
	defer rows.Close()
	var users []User
	for rows.Next() {
		var user User
		if err = rows.Scan(
			&user.Id,
			&user.Username,
			&user.Firstname,
			&user.Patronym,
			&user.Surname,
			&user.Rt,
		); err != nil {
			return users, err
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return users, err
	}
	return users, nil
}

func RpcGetUsers(c *gin.Context) {
	var data GetQuery
	err := c.BindJSON(&data)
	Unwrap(err)

	users, err := getUsers(data)
	Unwrap(err)

	c.JSON(200, users)
}

func setupDb(driver string, url string) {
	if driver == "" {
		driver = DEFAULT_DB_DRIVER
	}
	if url == "" {
		url = DEFAULT_DB_URL
	}

	dbDriver = driver
	_db, err := sql.Open(
		driver,
		url,
	)
	Unwrap(err)

	db = _db
}

func Assert(condition bool, msg ...string) {
	if !condition {
		// TODO: Allow `msg ...any`
		joined := strings.Join(msg, " ;; ")
		panic("Assertion Error: " + joined)
	}
}

type NewServerArgs struct {
	dbDriver string
	dbUrl    string
}

func newServer(args NewServerArgs) *gin.Engine {
	setupDb(args.dbDriver, args.dbUrl)
	server := gin.New()
	server.Use(gin.Recovery())

	server.POST("/rpc/login", RpcLogin)
	server.POST("/rpc/logout", RpcLogout)
	server.POST("/rpc/current", RpcCurrent)
	server.POST("/rpc/access", RpcAccess)

	server.POST("/rpc/server/reg", RpcReg)
	server.POST("/rpc/server/dereg", RpcDereg)
	server.POST("/rpc/server/get_users", RpcGetUsers)
	return server
}

func main() {
	server := newServer(NewServerArgs{})
	server.Run("localhost:9014")
}
