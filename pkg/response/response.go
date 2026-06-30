// pkg/response/response.go
//
// Standard response builder.
// SEMUA handler wajib pakai fungsi dari sini supaya format konsisten.
package response

// ErrorCode adalah tipe untuk kode error pada response.
type ErrorCode string

// Response adalah struktur utama semua response API.
type Response struct {
	Status     bool        `json:"status"`
	Data       any         `json:"data,omitempty"`
	ListData   any         `json:"list_data,omitempty"`
	Message    string      `json:"message"`
	Error      any         `json:"error,omitempty"`
	ErrorCode  ErrorCode   `json:"error_code,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Pagination berisi info halaman untuk response list.
type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}
