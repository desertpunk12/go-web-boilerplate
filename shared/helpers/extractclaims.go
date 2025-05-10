package helpers

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

func ExtractClaimsStr(key string, tokenRaw any) (string, error) {
	token, ok := tokenRaw.(*jwt.Token)
	if !ok {
		return "", fmt.Errorf("error converting tokenRaw to type *jwt.Token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("error token extracting and converting to jwt.MapClaims")
	}
	val, ok := claims[key].(string)
	if !ok || val == "" {
		return "", fmt.Errorf("error getting claims with key %s, got %v", key, val)
	}

	return val, nil
}
