package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	Header = "Authorization"
)

type TokenDetails struct {
	JTI string `json:"jti"`
	EXP int64  `json:"exp"`
}

func DecodeJWT(token string) (map[string]interface{}, error) {
	segments := strings.Split(token, ".")
	if len(segments) != 3 {
		return nil, fmt.Errorf("token contains an invalid number of segments")
	}

	payloadSeg, err := base64.RawURLEncoding.DecodeString(segments[1])
	if err != nil {
		return nil, err
	}

	var payload map[string]interface{}
	err = json.Unmarshal(payloadSeg, &payload)
	if err != nil {
		return nil, err
	}

	return payload, nil
}
