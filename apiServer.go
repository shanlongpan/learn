package main

import (
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
)

func main() {
	type Test struct {
		Name  string   `json:"name"`
		ID    int      `json:"id"`
		Age   int      `json:"age"`
		Title []string `json:"title"`
	}

	gin.DisableConsoleColor()

	// Logging to a file.
	f, _ := os.Create("./gin.log")
	gin.DefaultWriter = io.MultiWriter(f)

	r := gin.Default()
	r.POST("/ping", func(c *gin.Context) {
		var json Test
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if json.Name != "xiao" {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"name": "xiao", "id": 12})
	})
	r.Run("127.0.0.1:8091") // 监听8091端口
}
