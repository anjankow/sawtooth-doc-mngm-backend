package user

import (
	"doc-management/internal/config"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAppToken(t *testing.T) {
	viper.SetConfigFile("./../../.env")
	err := viper.ReadInConfig()
	require.NoError(t, err)

	manager := UserManager{
		tenantID: config.GetTenantID(),
		secret:   config.GetAppSecret(),
		clientID: config.GetClientID(),
	}

	token, err := manager.GetAppToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	t.Log(token)
}
