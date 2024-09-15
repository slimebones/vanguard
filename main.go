package main

import (
	"database/sql"
	"errors"
	"fmt"
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
type GetQuery = map[string]any

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
		Print(err)
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

func RpcLogout(c *gin.Context) {
	c.JSON(200, gin.H{})
}

func RpcCurrent(c *gin.Context) {
	c.JSON(200, gin.H{})
}

func RpcAccess(c *gin.Context) {
	c.JSON(200, gin.H{})
}

func RpcReg(c *gin.Context) {
	c.JSON(200, gin.H{})
}

func RpcDereg(c *gin.Context) {
	c.JSON(200, gin.H{})
}

func getUsers(gq GetQuery) {
}

func RpcGetUsers(c *gin.Context) {
	c.JSON(200, gin.H{})
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

func Assert(condition bool) {
	if !condition {
		panic("assertion error")
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
