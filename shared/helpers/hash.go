package helpers

import (
	"bytes"
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	"fmt"
)

func HashText(text string) (string, error) {
	hash := md5.New()
	_, err := hash.Write([]byte(text))
	if err != nil {
		return "", fmt.Errorf("error hashing password err: %v", err)
	}
	hashedText := hex.EncodeToString(hash.Sum(nil))

	return hashedText, nil
}

func HashAny(hashThis any) ([16]byte, error) {
	var gobBuffer bytes.Buffer
	encoder := gob.NewEncoder(&gobBuffer)
	err := encoder.Encode(hashThis)
	if err != nil {
		return [16]byte{}, nil
	}

	return md5.Sum(gobBuffer.Bytes()), nil
}
