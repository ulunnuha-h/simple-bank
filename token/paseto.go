package token

import (
	"fmt"
	"time"

	"github.com/o1egl/paseto"
)

type PasetoGenerator struct{
	secretkey string
}

func NewPasetoGenerator(secretkey string) (Generator, error){
	if len(secretkey) < 32 {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
	}

	return &PasetoGenerator{secretkey: secretkey}, nil
}

func (generator *PasetoGenerator) CreateToken(username string, duration time.Duration) (*Payload, string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return nil, "", err
	}
	token, err := paseto.NewV2().Encrypt([]byte(generator.secretkey), payload, payload.Footer)

	return payload, token, err
}

func (generator *PasetoGenerator) Verify(token string) (*Payload, error){
	var payload Payload
	err := paseto.NewV2().Decrypt(token, []byte(generator.secretkey), &payload, &payload.Footer)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if !payload.ExpiredAt.After(time.Now()) {
		return nil, ErrExpiredToken
	}

	return &payload, nil
}