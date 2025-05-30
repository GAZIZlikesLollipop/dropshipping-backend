package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

type Product struct {
	Id    int64  `json:"id"`
	Name  string `json:"name" binding:"required"`
	Price int    `json:"price" binding:"required"`
	Image string `json:"image" binding:"required"`
}

type User struct {
	Id        int64   `json:"id"`
	Name      string  `json:"name" binding:"required"`
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	Is_card   bool    `json:"is_card"`
	Cart      []int64 `json:"cart"`
}

var db *sql.DB

func main() {
	r := gin.Default()
	var err error
	db, err = sql.Open("sqlite3", "shop.db")
	if err != nil {
		log.Fatal("Ошибка создания базы данных")
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("Ошибка подключения базы данных")
	}

	productTable := `
		CREATE TABLE products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			price INTEGER,
			image TEXT NOT NULL
		)
	`
	_, err = db.Exec(productTable)
	if err != nil {
		log.Fatal("Ошибка создания таблицы продуктов")
	}

	userTable := `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			latitude REAL,
			longitude REAL,
			is_card INTEGER DEFAULT 0,
			cart TEXT 
		)
	`
	_, err = db.Exec(userTable)
	if err != nil {
		log.Fatal("Ошибка создания таблицы пользователей")
	}

	r.GET("/products", getProducts)
	r.GET("/product/:id", getProduct)
	r.DELETE("/product/:id", deleteProduct)
	r.POST("/product", addProduct)
	r.PATCH("/product/:id", updateProduct)

	r.GET("/users", getUsers)
	r.GET("/user/:id", getUser)
	r.DELETE("/user/:id", deleteUser)
	r.POST("/user", addUser)
	r.PATCH("/user/:id", updateUser)

	r.Run(":8080")
}
