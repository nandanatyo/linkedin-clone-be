package middleware

import (
	"linked-clone/pkg/response"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func FileUploadMiddleware(maxFileSize int64, allowedTypes []string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {

		if !strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
			c.Next()
			return
		}

		if err := c.Request.ParseMultipartForm(maxFileSize); err != nil {
			if strings.Contains(err.Error(), "request body too large") {
				response.Error(c, http.StatusRequestEntityTooLarge, "File too large", "File size exceeds maximum allowed size")
			} else {
				response.Error(c, http.StatusBadRequest, "Invalid multipart form", err.Error())
			}
			c.Abort()
			return
		}

		if c.Request.MultipartForm != nil && c.Request.MultipartForm.File != nil {
			for fieldName, files := range c.Request.MultipartForm.File {
				for _, fileHeader := range files {
					if !isAllowedFileType(fileHeader.Filename, allowedTypes) {
						response.Error(c, http.StatusBadRequest, "Invalid file type",
							"File type not allowed for field: "+fieldName)
						c.Abort()
						return
					}
				}
			}
		}

		c.Next()
	})
}

func isAllowedFileType(filename string, allowedTypes []string) bool {
	if len(allowedTypes) == 0 {
		return true
	}

	filename = strings.ToLower(filename)
	for _, allowedType := range allowedTypes {
		if strings.HasSuffix(filename, strings.ToLower(allowedType)) {
			return true
		}
	}
	return false
}
