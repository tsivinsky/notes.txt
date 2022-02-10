package main

import (
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("views/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{
			"message": "Write notes like it's 2000 again!",
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = ":5000"
	}
	r.Run(port)
}
