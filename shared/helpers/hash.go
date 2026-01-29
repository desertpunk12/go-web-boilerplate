package helpers

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"

	"encoding/gob"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashText hashes text using MD5. Note: This is NOT suitable for password hashing.
// For password hashing, use HashPass instead. This function is intended for
// generating simple hashes for non-sensitive purposes.
func HashText(text string) (string, error) {
	hash := md5.New()
	_, err := hash.Write([]byte(text))
	if err != nil {
		return "", fmt.Errorf("error hashing text err: %v", err)
	}
	hashedText := hex.EncodeToString(hash.Sum(nil))

	return hashedText, nil
}

// HashPass securely hashes a plaintext password using bcrypt with a default cost factor.
// bcrypt is specifically designed for password hashing and includes built-in salting
// and a configurable work factor to protect against brute-force attacks.
// The returned hash includes the salt and can be verified with CompareHashAndPassword.
func HashPass(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), 17)
	if err != nil {
		return "", fmt.Errorf("error hashing password err: %v", err)
	}
	return string(hashedBytes), nil
}

// CompareHashAndPassword securely compares a bcrypt hashed password with its possible
// plaintext equivalent. This function uses bcrypt's built-in timing-safe comparison
// to prevent timing attacks. Returns nil on success, or an error on failure.
// This is the correct way to verify passwords hashed with HashPass.
func CompareHashAndPassword(hashedPassword string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// HashAny computes a SHA256 hash of any Go struct/object using GOB encoding.
// Note: This is NOT suitable for password hashing - use HashPass instead.
// This function is intended for generating unique identifiers or checksums
// for non-sensitive data.
func HashAny(hashThis any) ([32]byte, error) {
	var gobBuffer bytes.Buffer
	encoder := gob.NewEncoder(&gobBuffer)
	err := encoder.Encode(hashThis)
	if err != nil {
		return [32]byte{}, nil
	}

	return sha256.Sum256(gobBuffer.Bytes()), nil
}
