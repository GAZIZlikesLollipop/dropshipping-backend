package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

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
		c.JSON(http.StatusNotFound, gin.H{"error": "пользователь не найден"})
		return
	} else if err != nil {
		log.Printf("Ошибка при получении пользовательа по ID %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении пользовательа: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func addUser(c *gin.Context) {
	var user User
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec(
		"INSERT INTO users (name,latitude,longitude,is_card,cart) VALUES (?,?,?,?,?)",
		user.Name, user.Latitude, user.Longitude, user.Is_card, "",
	)
	if err != nil {
		log.Println("Ошибка добавления пользователя в базу данных")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка добалвения пользователя в базу данных"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Пользователь успешно добавлен", "user": user})
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
		log.Printf("Ошибка при удалении пользовательа из базы данных: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении пользовательа из базы данных"})
		return
	}

	rowsAffectd, err := result.RowsAffected()
	if err != nil {
		log.Printf("Ошибка получения количества затронутых строк при удалении: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении пользовательа"})
		return
	}

	if rowsAffectd == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "пользователь не был найден"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Пользовтель успешно удален!"})
}

func updateUser(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println("Ошибка преоброзования пармтера")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка преоброзования пармтера"})
		return
	}

	var user User
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var currentUser User
	row := db.QueryRow("SELECT id,name,price,latitude,longitude,is_card,cart FROM users WHERE id = ?", id)
	err = row.Scan(&currentUser.Id, &currentUser.Name, &currentUser.Latitude, &currentUser.Longitude, &currentUser.Is_card, &currentUser.Cart)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	} else if err != nil {
		log.Printf("Ошибка при получении текущих данных пользователя с ID %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера при получении данных пользователя"})
		return
	}

	var (
		updateFields []string
		updateValues []interface{}
	)

	if user.Name != "" && currentUser.Name != user.Name {
		updateFields = append(updateFields, "name = ?")
		updateValues = append(updateValues, user.Name)
	}
	if user.Latitude != 0 && currentUser.Latitude != user.Latitude {
		updateFields = append(updateFields, "latitude = ?")
		updateValues = append(updateValues, user.Latitude)
	}
	if user.Longitude != 0 && currentUser.Longitude != currentUser.Longitude {
		updateFields = append(updateFields, "longitude = ?")
		updateValues = append(updateValues, user.Longitude)
	}
	if currentUser.Is_card != user.Is_card {
		updateFields = append(updateFields, "is_card = ?")
		updateValues = append(updateValues, user.Is_card)
	}
	if len(user.Cart) != 0 {
		updateFields = append(updateFields, "cart = ?")
		updateValues = append(updateValues, strings.Join(convertInt64ToStringSlice(user.Cart), ", "))
	}

	if len(updateFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Нету данных для обнвления"})
		return
	}

	updateQuery := fmt.Sprintf("UPDATE users SET %s WHERE id = ?", strings.Join(updateFields, ", "))
	updateValues = append(updateValues, id)

	stmt, err := db.Prepare(updateQuery)
	if err != nil {
		log.Println("Ошибка подготовки sql запроса")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка подготовки sql запроса"})
		return
	}
	defer stmt.Close()

	result, err := stmt.Exec(updateValues...)
	if err != nil {
		log.Printf("Ошибка при обновлении пользователья в базе данных: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении пользовательа в базе данных"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Ошибка получения количества затронутых строк при обновлении: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении пользователья"})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "пользователь не найден, и данные не измнеились"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Пользователь успешно обновлен", "user": user})
}
