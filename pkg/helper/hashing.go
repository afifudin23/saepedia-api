package helper

import "github.com/alexedwards/argon2id"

// Hash menghasilkan hash argon2id dari password plaintext.
func Hash(plain string) (string, error) {
	return argon2id.CreateHash(plain, argon2id.DefaultParams)
}

// Verify mencocokkan password plaintext dengan hash tersimpan.
func Verify(plain string, hashed string) bool {
	match, err := argon2id.ComparePasswordAndHash(plain, hashed)
	if err != nil {
		return false
	}
	return match
}
