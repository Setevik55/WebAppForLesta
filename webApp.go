package main

import (
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

type Word struct {
	Word string
	TF   int
	IDF  float64
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := initializeRouter()
	router.Run(":10001")
}

// Инициализация маршрутизатора
func initializeRouter() *gin.Engine {
	router := gin.New()
	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*")
	router.GET("/", openMainPage)
	router.POST("/upload", fileUpload)
	return router
}

// Обработка GET-запроса для открытия главной страницы
func openMainPage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

// Обработка POST-запроса для загрузки файла
func fileUpload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		showError(c, "Ошибка при получении файла")
		return
	}
	if !isValidFileType(file) {
		showError(c, "Недопустимый формат файла. Пожалуйста, загрузите текстовый документ (.txt)")
		return
	}
	inputBytes, err := openAndReadFile(file)
	if err != nil {
		showError(c, "Ошибка при чтении файла")
		return
	}
	words := extractWords(inputBytes)
	result := calcWordsAndTFIDF(words)
	result = sortDescIDF(result)

	c.HTML(http.StatusOK, "result.html", gin.H{"Words": result})
}

// Отправка ошибок в шаблон
func showError(c *gin.Context, message string) {
	c.HTML(http.StatusBadRequest, "error.html", gin.H{"Message": message})
}

// Проверка допустимого формата файла
func isValidFileType(file *multipart.FileHeader) bool {
	contentType := file.Header.Get("Content-Type")
	return strings.HasPrefix(contentType, "text/")
}

// Открытие и чтение файла
func openAndReadFile(file *multipart.FileHeader) ([]byte, error) {
	data, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer data.Close()

	inputBytes, err := io.ReadAll(data)
	if err != nil {
		return nil, err
	}
	return inputBytes, nil
}

// Разделение на слова и удаление лишних символов
func extractWords(input []byte) []string {
	stringWords := string(input)
	stringWords = strings.ToLower(stringWords)
	re := regexp.MustCompile(`(?:[a-zа-я]{2,}(?:-[a-zа-я]{2,})*)`)
	extractedWords := re.FindAllString(stringWords, -1)
	return extractedWords
}

// Подсчёт слов, tf и idf
func calcWordsAndTFIDF(words []string) []Word {
	countWords := make(map[string]int)
	for _, w := range words {
		countWords[w]++
	}
	idfMap := make(map[string]float64)
	for word := range countWords {
		idfMap[word] = math.Round(math.Log(float64(len(words))/float64(countWords[word]))*100) / 100
	}
	result := make([]Word, 0, len(countWords))
	for word, tf := range countWords {
		idf := idfMap[word]
		result = append(result, Word{word, tf, idf})
	}
	return result
}

// Сортировка списка по убыванию IDF
func sortDescIDF(WordsAndTFIDF []Word) []Word {
	n := len(WordsAndTFIDF)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if WordsAndTFIDF[j].IDF < WordsAndTFIDF[j+1].IDF {
				WordsAndTFIDF[j], WordsAndTFIDF[j+1] = WordsAndTFIDF[j+1], WordsAndTFIDF[j]
			}
		}
	}
	if len(WordsAndTFIDF) > 50 {
		return WordsAndTFIDF[:50] // Срезаем до первых 50
	}
	return WordsAndTFIDF
}
