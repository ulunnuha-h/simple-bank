package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/ulunnuha-h/simple_bank/util"
)

func CreateRandomUser(t *testing.T) User {
	hashedPassword, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)

	args := CreateUserParams{
		Username: util.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName: util.RandomString(5),
		Email: util.RandomEmail(),
	}

	user, err := testQuery.CreateUser(context.Background(), args)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, args.Email, user.Email)
	require.Equal(t, args.FullName, user.FullName)
	require.Equal(t, args.HashedPassword, user.HashedPassword)
	require.Equal(t, args.Username, user.Username)

	require.NotEmpty(t, user.CreatedAt)
	require.NotEmpty(t, user.PasswordChangedAt)

	return user
}

func TestCreateUser(t *testing.T) {
	CreateRandomUser(t)
}

func TestGetuser(t *testing.T) {
	user := CreateRandomUser(t)
	user2, err := testQuery.GetUser(context.Background(), user.Username)

	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user.Email, user2.Email)
	require.Equal(t, user.FullName, user2.FullName)
	require.Equal(t, user.HashedPassword, user2.HashedPassword)
	require.Equal(t, user.Username, user2.Username)
	require.WithinDuration(t, user.CreatedAt, user2.CreatedAt, time.Second)
	require.WithinDuration(t, user.PasswordChangedAt, user2.PasswordChangedAt, time.Second)
}