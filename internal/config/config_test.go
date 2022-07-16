package config

import (
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestTimeout(t *testing.T) {
	viper.Set("REQ_TIMEOUT", "")
	timeout := GetRequestTimeout()
	assert.Equal(t, timeout, defaultRequestTimeout)

	viper.Set("REQ_TIMEOUT", "14s")
	timeout = GetRequestTimeout()
	assert.Equal(t, timeout, 14*time.Second)
}
