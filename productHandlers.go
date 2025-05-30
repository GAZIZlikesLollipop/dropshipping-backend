package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
		if err := rows.Scan(&p.Id, &p.Name, &p.Price, &p.Image); err != nil {
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

	err = row.Scan(&product.Id, &product.Name, &product.Price, &product.Image)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Продукт не найден"})
		return
	} else if err != nil {
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
		Image: imageUrl,
	}

	stmt, err := db.Prepare("INSERT INTO products(name,price,image) VALUES(?,?,?)")
	if err != nil {
		log.Printf("Ошибка подготовки SQL-запроса: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка подготовки SQL-запроса"})
		return
	}
	defer stmt.Close()

	result, err := stmt.Exec(product.Name, product.Price, product.Image)
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
	idStr := c.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println("Ошибка преоброзования пармтера")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка преоброзования пармтера"})
		return
	}

	var currentProduct Product
	row := db.QueryRow("SELECT id,name,price,image FROM products WHERE id = ?", id)
	err = row.Scan(&currentProduct.Id, &currentProduct.Name, &currentProduct.Price, &currentProduct.Image)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Продукт не найден"})
		return
	} else if err != nil {
		log.Printf("Ошибка при получении текущих данных продукта с ID %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера при получении данных продукта"})
		return
	}
	newName := c.PostForm("name")
	newPriceStr := c.PostForm("price")
	newImageFile, fileError := c.FormFile("image")

	var updateFields []string
	var updateValues []interface{}

	if newName != "" && newName != currentProduct.Name {
		updateFields = append(updateFields, "name = ?")
		updateValues = append(updateValues, newName)
		currentProduct.Name = newName
	}

	if newPriceStr != "" {
		newPrice, priceErr := strconv.Atoi(newPriceStr)
		if priceErr != nil {
			log.Println("Ошибка парсинга новой цены")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка парсинга новой цены"})
			return
		}
		if newPrice <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Цена должна быть больше нуля"})
			return
		}
		if newPrice != currentProduct.Price {
			updateFields = append(updateFields, "price = ?")
			updateValues = append(updateValues, newPrice)
			currentProduct.Price = newPrice
		}
	}

	// Обновление изображения продукта
	// Проверяем, был ли файл "image" отправлен в запросе (`fileHeaderErr == nil` означает, что файл успешно получен).
	if fileError == nil && newImageFile != nil {
		// --- Логика удаления старого изображения ---
		if currentProduct.Image != "" && currentProduct.Image != "/" {
			filePathOnDisk := filepath.Join(".", currentProduct.Image)
			// Важная проверка безопасности: убедитесь, что удаляете только из вашей папки загрузок.
			// Это предотвратит попытки удалить файлы из системных директорий.
			if strings.HasPrefix(filePathOnDisk, "uploads"+string(filepath.Separator)) {
				if err := os.Remove(filePathOnDisk); err != nil {
					log.Printf("Ошибка при удалении старого файла изображения %s: %v", filePathOnDisk, err)
					// В реальном приложении можно было бы логировать или отправлять предупреждение.
					// Здесь мы продолжаем, так как обновление записи в БД важнее, чем удаление старого файла.
				} else {
					log.Printf("Старый файл %s успешно удален с диска", filePathOnDisk)
				}
			} else {
				log.Printf("Попытка удалить файл вне директории загрузок: %s (пропускаем удаление старого файла)", filePathOnDisk)
			}
		}

		// --- Логика сохранения нового изображения ---
		// Генерируем уникальное имя файла с помощью метки времени и оригинального имени
		newFilename := fmt.Sprintf("%d-%s", time.Now().UnixNano(), filepath.Base(newImageFile.Filename))
		uploadDir := filepath.Join(".", "uploads", "images")   // Директория, куда будем сохранять
		newUploadPath := filepath.Join(uploadDir, newFilename) // Полный путь к новому файлу

		// Убедимся, что директория для загрузок существует (она уже создается в setupDatabase, но дополнительная проверка не помешает)
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			log.Printf("Ошибка создания директории для загрузки '%s': %v", uploadDir, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера при сохранении нового изображения"})
			return
		}

		// Сохраняем загруженный файл на диске
		if err := c.SaveUploadedFile(newImageFile, newUploadPath); err != nil {
			log.Printf("Ошибка сохранения нового файла '%s': %v", newUploadPath, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения нового изображения"})
			return
		}

		// Формируем URL изображения для сохранения в базе данных
		// Этот URL будет использоваться клиентом для доступа к изображению через статический сервер
		newImageUrl := "/" + filepath.Join("uploads", "images", newFilename)
		updateFields = append(updateFields, "image_url = ?") // Добавляем поле для обновления в SQL
		updateValues = append(updateValues, newImageUrl)     // Добавляем значение
		currentProduct.Image = newImageUrl                   // Обновляем в объекте для возврата клиенту
	}

	if len(updateFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Нету данных для обнвления"})
		return
	}

	updateQuery := fmt.Sprintf("UPDATE products SET %s WHERE id = ?", strings.Join(updateFields, ", "))
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
		log.Printf("Ошибка при обновлении продукта в базе данных: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении продукта в базе данных"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Ошибка получения количества затронутых строк при обновлении: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении продукта"})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Продукт не найден, и данные не измнеились"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Данные продукта успешно обновленны\n%v", currentProduct)})
}
