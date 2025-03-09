package token

import "time"

type Generator interface{
	CreateToken(username string, duration time.Duration) (string, error)
	Verify(token string) (*Payload, error)
}