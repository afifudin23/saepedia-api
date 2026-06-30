package config

// Version adalah versi rilis aplikasi (dipakai release manager & /ping).
const Version = "0.1.0"

// APIVersion adalah versi PATH API (segmen URL), mis. "v1" → /api/v1 & /docs/v1.
// Beda dengan Version (semver rilis) di atas. Untuk naik versi: ubah ke "v2",
// samakan @BasePath di cmd/api/main.go, lalu `make swag`.
const APIVersion = "v1"

// APIBasePath mengembalikan prefix route, mis. "/api/v1".
func APIBasePath() string { return "/api/" + APIVersion }
