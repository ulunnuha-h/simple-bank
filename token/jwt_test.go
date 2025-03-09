package token

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/ulunnuha-h/simple_bank/util"
)

func TestGenerateJWTToken(t *testing.T){
	generator, err := NewJWTGenerator(util.RandomString(32))
	require.NoError(t, err)

	username := util.RandomOwner()
	duration := time.Minute

	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)

	token, err := generator.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := generator.Verify(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	require.NotZero(t, payload.ID)
	require.Equal(t, payload.Username, username)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
}

func TestExpiredToken(t *testing.T){
	generator, err := NewJWTGenerator(util.RandomString(32))
	require.NoError(t, err)

	token, err := generator.CreateToken(util.RandomOwner(), -time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := generator.Verify(token)
	
	require.Error(t, err)
	require.EqualError(t, err, ErrEpiredToken.Error())
	require.Nil(t, payload)
}