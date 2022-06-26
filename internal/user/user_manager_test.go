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

func TestGetUserByIDWithKeys(t *testing.T) {
	manager := newTestUserManager(t)
	userID := "05c0f0f3-65dc-4a88-8b13-1aa4995ff4c3"
	user, err := manager.GetUserByID(userID)
	require.NoError(t, err)
	assert.Equal(t, "test@csunivie3.onmicrosoft.com", user.Name)
	t.Log(t, user.PrivateKey)
	t.Log(t, user.PublicKey)
}

func TestGetUserByIDWithoutKeys(t *testing.T) {
	manager := newTestUserManager(t)
	userID := "988c6996-da05-43bb-811b-4c7d78f046fa"
	user, err := manager.GetUserByID(userID)
	require.NoError(t, err)
	assert.Equal(t, "test2@csunivie3.onmicrosoft.com", user.Name)
	t.Log(t, user.PrivateKey)
	t.Log(t, user.PublicKey)
}
