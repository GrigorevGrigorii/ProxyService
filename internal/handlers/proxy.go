package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Ping(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, gin.H{"status": "ok"})
}

func ProxyGetRequest(c *gin.Context) {
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}

func ProxyPostRequest(c *gin.Context) {
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}

func ProxyPutRequest(c *gin.Context) {
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}

func ProxyDeleteRequest(c *gin.Context) {
	c.IndentedJSON(http.StatusNotImplemented, gin.H{"message": "not_implemented_error"})
}
