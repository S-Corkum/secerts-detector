package main

import (
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"secrets-detector/pkg/models"
)

func validateCertificate(cert string) bool {
	block, _ := pem.Decode([]byte(cert))
	if block == nil {
		return false
	}

	_, err := x509.ParseCertificate(block.Bytes)
	return err == nil
}

func validatePrivateKey(key string) bool {
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return false
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		_, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		return err == nil
	case "EC PRIVATE KEY":
		_, err := x509.ParseECPrivateKey(block.Bytes)
		return err == nil
	}
	return false
}

func validateAWSKey(key string) bool {
	return strings.HasPrefix(key, "AKIA") && len(key) == 20
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.POST("/validate", func(c *gin.Context) {
		var req models.ValidationRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ValidationResponse{
				IsValid: false,
				Message: "Invalid request format",
			})
			return
		}

		var isValid bool
		var message string

		switch req.Secret.Type {
		case "certificate":
			isValid = validateCertificate(req.Secret.Value)
			message = "Invalid certificate format"
		case "private_key":
			isValid = validatePrivateKey(req.Secret.Value)
			message = "Invalid private key format"
		case "aws_key":
			isValid = validateAWSKey(req.Secret.Value)
			message = "Invalid AWS key format"
		default:
			isValid = false
			message = "Unknown secret type"
		}

		if isValid {
			message = "Valid " + req.Secret.Type
		}

		c.JSON(http.StatusOK, models.ValidationResponse{
			IsValid: isValid,
			Message: message,
		})
	})

	return r
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := setupRouter()
	r.Run(":" + port)
}
