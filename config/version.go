package config

// Version adalah versi rilis aplikasi (dipakai release manager & /ping).
const Version = "0.0.0"

// APIVersion adalah versi path API (prefix /api/<APIVersion>).
// Ganti di sini untuk menaikkan versi; samakan juga @BasePath di cmd/api/main.go
// lalu jalankan `make swag`.
const APIVersion = "v1"

// APIBasePath mengembalikan prefix route, mis. "/api/v1".
func APIBasePath() string { return "/api/" + APIVersion }
