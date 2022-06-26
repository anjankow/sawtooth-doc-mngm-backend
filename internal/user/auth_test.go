package user

import (
	"doc-management/internal/config"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetAppToken(t *testing.T) {
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

	firstToken := manager.tokenGuard.token

	assert.NotEmpty(t, firstToken)
	t.Log(firstToken)

}
