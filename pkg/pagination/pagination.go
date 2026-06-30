package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	defaultPage    = 1
	defaultPerPage = 10
	maxPerPage     = 100
)

// Params hasil parse query ?page=&per_page=.
type Params struct {
	Page    int
	PerPage int
}

// Offset menghitung offset untuk query database.
func (p Params) Offset() int { return (p.Page - 1) * p.PerPage }

// Parse membaca page & per_page dari query string dengan default & batas aman.
func Parse(c *gin.Context) Params {
	page := atoiDefault(c.Query("page"), defaultPage)
	if page < 1 {
		page = defaultPage
	}

	perPage := atoiDefault(c.Query("per_page"), defaultPerPage)
	if perPage < 1 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}

	return Params{Page: page, PerPage: perPage}
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}
