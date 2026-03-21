package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func PingHandler(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, gin.H{"status": "ok"})
}

func GetProxyHandler(c *gin.Context) {
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}

func PostProxyHandler(c *gin.Context) {
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}

func PutProxyHandler(c *gin.Context) {
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}

func DeleteProxyHandler(c *gin.Context) {
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}
