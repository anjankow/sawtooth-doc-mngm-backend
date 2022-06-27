package usermanager

import (
	"context"
	"doc-management/internal/config"
	"doc-management/internal/model"
	"doc-management/internal/signkeys"
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

func TestReadAppKeys(t *testing.T) {

	manager := newTestUserManager(t)
	appUserID := "1ce7cffe-bf4a-4a14-8535-413013cae16f"
	keys, err := manager.InitAndReadAppKeys(context.TODO(), appUserID)
	require.NoError(t, err)
	assert.NotEmpty(t, keys.PrivateKey.AsHex())
	assert.NotEmpty(t, keys.PublicKey.AsHex())
	t.Log(t, keys.PublicKey.AsHex())

	pub1 := keys.PublicKey.AsHex()
	keys2, err := manager.InitAndReadAppKeys(context.TODO(), appUserID)
	pub2 := keys2.PublicKey.AsHex()
	assert.Equal(t, pub1, pub2)
}

func TestUpdateUserKeys(t *testing.T) {
	keys, err := signkeys.GenerateKeys()
	require.NoError(t, err)

	manager := newTestUserManager(t)
	user := model.User{
		ID:   "05c0f0f3-65dc-4a88-8b13-1aa4995ff4c3",
		Name: "test@csunivie3.onmicrosoft.com",
	}
	user, err = manager.updateUserKeys(context.TODO(), user, keys)
	require.NoError(t, err)
	assert.Equal(t, "test@csunivie3.onmicrosoft.com", user.Name)
	assert.Equal(t, keys.PrivateKey.AsHex(), user.Keys.PrivateKey.AsHex())
	assert.Equal(t, keys.PublicKey.AsHex(), user.Keys.PublicKey.AsHex())

	updated, err := manager.getUserByID(context.TODO(), user.ID)
	require.NoError(t, err)
	assert.Equal(t, "test@csunivie3.onmicrosoft.com", updated.Name)
	assert.Equal(t, keys.PrivateKey.AsHex(), user.Keys.PrivateKey.AsHex())
	assert.Equal(t, keys.PublicKey.AsHex(), user.Keys.PublicKey.AsHex())

}

func TestGetUserByIDWithKeys(t *testing.T) {
	manager := newTestUserManager(t)
	userID := "05c0f0f3-65dc-4a88-8b13-1aa4995ff4c3"
	user, err := manager.getUserByID(context.TODO(), userID)
	require.NoError(t, err)
	assert.Equal(t, "test@csunivie3.onmicrosoft.com", user.Name)
	t.Log(t, user.Keys.PrivateKey)
	t.Log(t, user.Keys.PublicKey)
}

func TestGetUserByIDWithoutKeys(t *testing.T) {
	manager := newTestUserManager(t)
	userID := "988c6996-da05-43bb-811b-4c7d78f046fa"
	user, err := manager.getUserByID(context.TODO(), userID)
	require.NoError(t, err)
	assert.Equal(t, "test2@csunivie3.onmicrosoft.com", user.Name)
	t.Log(t, user.Keys.PrivateKey)
	t.Log(t, user.Keys.PublicKey)
}
