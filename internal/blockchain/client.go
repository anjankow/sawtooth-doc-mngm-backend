package blockchain

import "go.uber.org/zap"

type Client struct {
	logger *zap.Logger
	url    string
}

func NewClient(logger *zap.Logger, validatorRestAPIUrl string) *Client {
	return &Client{logger: logger, url: validatorRestAPIUrl}
}
