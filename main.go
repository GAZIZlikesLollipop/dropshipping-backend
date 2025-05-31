package main

import (
	"database/sql"
	"fmt"
	"log"
    "strings"
    "strconv"
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

func convertInt64ToStringSlice(intSlice []int64) []string {
	var stringSlice []string
	for _, i := range intSlice {
		stringSlice = append(stringSlice, fmt.Sprintf("%d", i))
	}
	return stringSlice
}

func parseCart(cart string) []int64 {
	var result []int64
	if cart == "" {
		return result
	}
	parts := strings.Split(cart, ",")
	for _, part := range parts {
		num, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			log.Println("Некорректные данные в cart:", part)
			continue
		}
		result = append(result, num)
	}
	return result
}

var db *sql.DB

func main() {
	r := gin.Default()
	var err error
	db, err = sql.Open("sqlite3", "shop.db")
	if err != nil {
		log.Fatalf("Ошибка создания базы данных\n%v",err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Ошибка подключения базы данных\n%v",err)
	}

	productTable := `
		CREATE TABLE IF NOT EXISTS products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			price INTEGER,
			image TEXT NOT NULL
		)
	`
	_, err = db.Exec(productTable)
	if err != nil {
		log.Fatalf("Ошибка создания таблицы продуктов\n%v",err)
	}

	userTable := `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			latitude REAL,
			longitude REAL,
			is_card INTEGER,
			cart TEXT 
		)
	`
	_, err = db.Exec(userTable)
	if err != nil {
		log.Fatalf("Ошибка создания таблицы пользователей\n%v",err)
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
