#!/bin/sh
set -e

# URL-encode karakter spesial di password (untuk format URL golang-migrate)
encode() {
  printf '%s' "$1" | sed \
    -e 's/%/%25/g' -e 's/@/%40/g' -e 's/:/%3A/g' \
    -e 's/\//%2F/g' -e 's/?/%3F/g' -e 's/#/%23/g' \
    -e 's/\[/%5B/g' -e 's/\]/%5D/g' -e 's/!/%21/g' \
    -e 's/\$/%24/g' -e 's/&/%26/g'
}

ENC_PASS=$(encode "$DB_PASSWORD")
DB_URL="postgres://${DB_USER}:${ENC_PASS}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

echo "==> Menjalankan migrasi database..."
migrate -path ./migrations -database "$DB_URL" up

echo "==> Migrasi selesai. Menjalankan aplikasi..."
exec "$@"
