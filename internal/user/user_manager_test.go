package user

import (
	"doc-management/internal/config"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestUserManager(t *testing.T) UserManager {
	viper.SetConfigFile("./../../.env")
	err := viper.ReadInConfig()
	require.NoError(t, err)

	manager, err := NewUserManager(
		config.GetTenantID(),
		config.GetClientID(),
		config.GetMsExtensionID(),
		config.GetAppSecret(),
	)
	require.NoError(t, err)

	return manager

}

func TestUpdateUserKeys(t *testing.T) {
	publicKey := "sfds"
	privateKey := "sfds22"
	manager := newTestUserManager(t)
	user := User{
		ID:   "05c0f0f3-65dc-4a88-8b13-1aa4995ff4c3",
		Name: "test@csunivie3.onmicrosoft.com",
	}
	user, err := manager.updateUserKeys(user, privateKey, publicKey)
	require.NoError(t, err)
	assert.Equal(t, "test@csunivie3.onmicrosoft.com", user.Name)
	assert.Equal(t, privateKey, user.PrivateKey)
	assert.Equal(t, publicKey, user.PublicKey)

	updated, err := manager.getUserByID(user.ID)
	require.NoError(t, err)
	assert.Equal(t, "test@csunivie3.onmicrosoft.com", updated.Name)
	assert.Equal(t, privateKey, updated.PrivateKey)
	assert.Equal(t, publicKey, updated.PublicKey)

}

func TestGetUserByIDWithKeys(t *testing.T) {
	manager := newTestUserManager(t)
	userID := "05c0f0f3-65dc-4a88-8b13-1aa4995ff4c3"
	user, err := manager.getUserByID(userID)
	require.NoError(t, err)
	assert.Equal(t, "test@csunivie3.onmicrosoft.com", user.Name)
	t.Log(t, user.PrivateKey)
	t.Log(t, user.PublicKey)
}

func TestGetUserByIDWithoutKeys(t *testing.T) {
	manager := newTestUserManager(t)
	userID := "988c6996-da05-43bb-811b-4c7d78f046fa"
	user, err := manager.getUserByID(userID)
	require.NoError(t, err)
	assert.Equal(t, "test2@csunivie3.onmicrosoft.com", user.Name)
	t.Log(t, user.PrivateKey)
	t.Log(t, user.PublicKey)
}
