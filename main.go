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

const DEFAULT_DB_DRIVER = "postgres"
const DEFAULT_DB_URL = "postgres://vanguard:vanguard@localhost:9005/vanguard"

type Id = int32
type Time = int64

func utc() Time {
	return time.Now().Unix()
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func newToken(secret string, userId Id) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userId,
		"created": utc(),
	})
	return token.SignedString(secret)
}

func Err(msg string) error {
	return errors.New(msg)
}

type Token struct {
	UserId  Id
	Created Time
}

func parseToken(token string, secret string) (*Token, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(
		token,
		claims,
		func(t *jwt.Token) (interface{}, error) { return []byte(secret), nil },
	)
	if err != nil {
		return nil, err
	}
	userId, ok := claims["user_id"].(Id)
	if !ok {
		return nil, Err("cannot parse user id")
	}
	created, ok := claims["created"].(Time)
	if !ok {
		return nil, Err("cannot parse creation time")
	}
	return &Token{userId, created}, nil
}

func createUser(
	username string,
	password string,
	firstname string,
	patronym string,
	surname string,
) {
	hpassword, err := hashPassword(password)
	Unwrap(err)
	rows, err := db.Query(
		`
			INSERT INTO appuser (
				username, hpassword, firstname, patronym, surname
			) VALUES
				($1, $2, $3, $4, $5)
			RETURNING *
		`,
		username,
		hpassword,
		firstname,
		patronym,
		surname,
	)
	Unwrap(err)
	Print(rows)
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

func login(c *gin.Context) {
	var data Login
	err := c.BindJSON(&data)
	Unwrap(err)

	// user := db.QueryRow(
	// 	`SELECT * FROM appuser WHERE username = $1`, data.Username,
	// )

	c.JSON(200, gin.H{"donuts": true})
}

func logout(c *gin.Context) {
	c.JSON(200, gin.H{})
}

func current(c *gin.Context) {
	c.JSON(200, gin.H{})
}

func access(c *gin.Context) {
	c.JSON(200, gin.H{})
}

func reg(c *gin.Context) {
	c.JSON(200, gin.H{})
}

func dereg(c *gin.Context) {
	c.JSON(200, gin.H{})
}

func get_users(c *gin.Context) {
	c.JSON(200, gin.H{})
}

func setupDb(driver string, url string) {
	if driver == "" {
		driver = DEFAULT_DB_DRIVER
	}
	if url == "" {
		url = DEFAULT_DB_URL
	}

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

	server.POST("/rpc/login", login)
	server.POST("/rpc/logout", logout)
	server.POST("/rpc/current", current)
	server.POST("/rpc/access", access)

	server.POST("/rpc/server/reg", reg)
	server.POST("/rpc/server/dereg", dereg)
	server.POST("/rpc/server/get_users", get_users)
	return server
}

func main() {
	server := newServer(NewServerArgs{})
	server.Run("localhost:9014")
}
