package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Success mengembalikan response sukses dengan single object.
func Success(c *gin.Context, statusCode int, message string, data any) {
	c.JSON(statusCode, Response{
		Status:  true,
		Data:    data,
		Message: message,
	})
}

// List mengembalikan response sukses dengan list data dan info pagination.
func List(c *gin.Context, page int, perPage int, total int, message string, listData any) {
	totalPages := 0
	if perPage > 0 {
		totalPages = (total + perPage - 1) / perPage
	}

	c.JSON(http.StatusOK, Response{
		Status:   true,
		ListData: listData,
		Message:  message,
		Pagination: &Pagination{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}
