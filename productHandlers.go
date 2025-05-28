package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func getProducts(c *gin.Context) {
	rows, err := db.Query("SELECT id,name,price,image FROM products")
	if err != nil {
		log.Println("Ошибка получения продуктов")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения продуктов"})
		return
	}
	defer rows.Close()
	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.Id, &p.Name, &p.Price, &p.Iamge); err != nil {
			log.Printf("Ошибка сканирования продукта: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Ошибка сканирования продукта: %v", err)})
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Ошибка итерации по продуктам: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка итерации по продуктам: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, products)
}

func getProduct(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println("Ошибка преоброзования пармтера")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка преоброзования пармтера"})
		return
	}
	row := db.QueryRow("SELECT id,name,price,image FROM products WHERE id = ?", id)
	var product Product
	err = row.Scan(&product.Id, &product.Name, &product.Price, &product.Iamge)
	if err == sql.ErrNoRows {
		// Если продукт с таким ID не найден
		c.JSON(http.StatusNotFound, gin.H{"error": "Продукт не найден"})
		return
	} else if err != nil {
		// Другая ошибка при сканировании или выполнении запроса
		log.Printf("Ошибка при получении продукта по ID %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении продукта: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, product)
}

func deleteProduct(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println("Ошибка преоброзования пармтера")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка преоброзования пармтера"})
		return
	}

	var imageUrl string
	rows := db.QueryRow("SELECT image FROM products WHERE id = ?", id)
	err = rows.Scan(&imageUrl)
	if err == sql.ErrNoRows {
		// Если продукта не существует, то и удалять нечего
		c.JSON(http.StatusNotFound, gin.H{"error": "Продукт не найден"})
		return
	} else if err != nil {
		log.Printf("Ошибка при получении image_url продукта с ID %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера при подготовке к удалению"})
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

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Ошибка получения количества затронутых строк при удалении: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении продукта"})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Продукт не был найден"})
		return
	}

	if imageUrl != "" && imageUrl != "/" {
		// Убедитесь, что "uploads" соответствует вашей базовой директории для статики
		// И imageUrl начинается с "/uploads/"
		filePathOnDisk := filepath.Join(".", imageUrl) // Используем "." чтобы получить относительный путь от корня приложения

		// Защита от удаления критически важных файлов
		if !filepath.HasPrefix(filePathOnDisk, "uploads"+string(filepath.Separator)) {
			log.Printf("Попытка удалить файл вне директории загрузок: %s", filePathOnDisk)
			// Не прерываем удаление из БД, но логируем проблему с файлом
		} else {
			err := os.Remove(filePathOnDisk)
			if err != nil {
				log.Printf("Ошибка при удалении файла %s: %v", filePathOnDisk, err)
				// В реальном приложении можно было бы отправить 200 OK, но с предупреждением о файле.
				// Сейчас просто логируем и продолжаем.
			} else {
				log.Printf("Файл %s успешно удален с диска", filePathOnDisk)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Подукт успешно удален!"})
}

func addProduct(c *gin.Context) {
	name := c.PostForm("name")
	priceStr := c.PostForm("price")

	imageFile, err := c.FormFile("image")
	if err != nil {
		log.Printf("Ошибка при получении файла изображения: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Ошибка при получении файла изображения: %v", err)})
		return
	}

	price, err := strconv.Atoi(priceStr)
	if err != nil {
		log.Printf("Ошибка парсинга цены '%s': %v", priceStr, err) // Логируем, что пришло
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверное значение цены"})
		return
	}

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Имя прдукта не введено"})
		return
	}
	if price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Цена прдукта не введено"})
		return
	}

	filename := fmt.Sprintf("%d-%s", time.Now().UnixNano(), filepath.Base(imageFile.Filename))

	uploadDir := filepath.Join(".", "uploads", "images") // Путь к директории загрузок от корня приложения
	uploadPath := filepath.Join(uploadDir, filename)     // Полный путь к файлу на диске

	if err := os.MkdirAll(uploadDir, 0755); err != nil { // 0755 - права доступа
		log.Printf("Ошибка создания директории для загрузки '%s': %v", uploadDir, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера при сохранении изображения"})
		return
	}

	if err := c.SaveUploadedFile(imageFile, uploadPath); err != nil {
		log.Printf("Ошибка сохранения файла '%s': %v", uploadPath, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения изображения"})
		return
	}

	imageUrl := "/" + filepath.Join("uploads", "images", filename)

	product := Product{
		Name:  name,
		Price: price,
		Iamge: imageUrl,
	}

	stmt, err := db.Prepare("INSERT INTO products(name,price,image) VALUES(?,?,?)")
	if err != nil {
		log.Printf("Ошибка подготовки SQL-запроса: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка подготовки SQL-запроса"})
		return
	}
	defer stmt.Close()

	result, err := stmt.Exec(product.Name, product.Price, product.Iamge)
	if err != nil {
		log.Printf("Ошибка при добавлении продукта в базу данных: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при добавлении продукта в базу данных"})
		return
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Ошибка получения ID нового продукта: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения ID нового продукта"})
		return
	}
	product.Id = id
	c.JSON(http.StatusCreated, gin.H{"message": fmt.Sprintf("Продукт успешно добавлен!\n%v", product)})
}

func updateProduct(c *gin.Context) {

}
