package blockchain

import (
	"strings"

	"go.uber.org/zap"
)

type Client struct {
	logger *zap.Logger
	url    string
}

func NewClient(logger *zap.Logger, validatorRestAPIUrl string) *Client {
	url := validatorRestAPIUrl
	if !strings.HasPrefix(validatorRestAPIUrl, "http://") {
		url = "http://" + validatorRestAPIUrl
	}

	return &Client{logger: logger, url: url}
}
