package token

import "time"

type Generator interface{
	CreateToken(username string, duration time.Duration) (*Payload, string, error)
	Verify(token string) (*Payload, error)
}