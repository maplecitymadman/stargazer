package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse sends a standardized error response
func ErrorResponse(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{"error": message})
}

// ErrorResponsef sends a formatted error response
func ErrorResponsef(c *gin.Context, code int, format string, args ...interface{}) {
	c.JSON(code, gin.H{"error": fmt.Sprintf(format, args...)})
}

// SuccessResponse sends a standardized success response with data
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// SuccessMessageResponse sends a success response with a message
func SuccessMessageResponse(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{"message": message, "success": true})
}

// ClientNotInitializedResponse sends a standard "client not initialized" error
func ClientNotInitializedResponse(c *gin.Context) {
	ErrorResponse(c, http.StatusServiceUnavailable,
		"Kubernetes client not initialized. Please configure kubeconfig.")
}

// BadRequestResponse sends a 400 Bad Request with the given message
func BadRequestResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusBadRequest, message)
}

// NotFoundResponse sends a 404 Not Found with the given message
func NotFoundResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusNotFound, message)
}

// InternalErrorResponse sends a 500 Internal Server Error with the given message
func InternalErrorResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusInternalServerError, message)
}

// InternalErrorResponsef sends a formatted 500 Internal Server Error
func InternalErrorResponsef(c *gin.Context, format string, args ...interface{}) {
	ErrorResponsef(c, http.StatusInternalServerError, format, args...)
}
