package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func getUsers(c *gin.Context) {
	rows, err := db.Query("SELECT id,name,latitude,longitude,is_card,cart FROM users")
	if err != nil {
		log.Println("Ошибка получения пользовательей")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения пользовательей"})
		return
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.Id, &u.Name, &u.Latitude, &u.Longitude, &u.Is_card, &u.Cart); err != nil {
			log.Printf("Ошибка сканирования пользователья: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Ошибка сканирования пользователья: %v", err)})
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Ошибка итерации по пользовательям: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка итерации по пользовательям: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

func getUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println("Ошибка преоброзования пармтера")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка преоброзования пармтера"})
		return
	}

	row := db.QueryRow("SELECT id,name,latitude,longitude,is_card,cart FROM users WHERE id = ?", id)

	var user User
	err = row.Scan(&user.Id, &user.Name, &user.Latitude, &user.Longitude, &user.Is_card, &user.Cart)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Продукт не найден"})
		return
	} else if err != nil {
		log.Printf("Ошибка при получении продукта по ID %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении продукта: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func addUser(c *gin.Context) {

}

func deleteUser(c *gin.Context) {

	idStr := c.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println("Ошибка преоброзования пармтера")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка преоброзования пармтера"})
		return
	}

	stmt, err := db.Prepare("DELETE FROM products WHERE id = ?")
	if err != nil {
		log.Printf("Ошибка подготовки SQL-запроса на удаление: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка подготовки SQL-запроса"})
		return
	}

	result, err := stmt.Exec(id)

	if err != nil {
		log.Printf("Ошибка при удалении продукта из базы данных: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении продукта из базы данных"})
		return
	}

	rowsAffectd, err := result.RowsAffected()
	if err != nil {
		log.Printf("Ошибка получения количества затронутых строк при удалении: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении продукта"})
		return
	}

	if rowsAffectd == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Продукт не был найден"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Пользовтель успешно удален!"})
}

func updateUser(c *gin.Context) {

}
