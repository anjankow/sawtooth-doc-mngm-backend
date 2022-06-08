package blockchain

import (
	"doc-management/internal/config"
	"doc-management/internal/hashing"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.uber.org/zap"
)

func TestGetPayload(t *testing.T) {
	logger := zap.NewExample()
	client := NewClient(logger, config.GetValidatorRestAPIAddr())
	hashing.Initialize(logger)

	responsePayload := `
	{
		"data": "oWlwcm9wb3NhbHOjeCQ4ZDZkMzg5MS01MGRkLTQxOGUtYjNhMS01YzZkMmQ4NTBmYzV4gDYzZmNiYzhmYzk0ZDBmZGVkMDM5ODMwMmZhMDdmYTRlNDQ5YTg2NjBkNmE1NWY2MTUwNGI4YzQ2Mzc0YTE2ZjlkMjgxZWM2ZTBhMzZkNWI5MGY0OGE1NDU0MzMxZWJmNzNkZTQ4MjQwNTQxYWU0MzQ1NTlhMTk0OWJiNjZjZTljeCRkMjVkMmY5Ni00ZWQ0LTQ3NDctYTA3YS05MDQ4NGI1ZjgyNmR4gDA5NTU2ZWM1NDRhYWIzYzczYmUxOGJjNDQ4NjIzMjY4OTg0ZTkxMzk2YjgzNjNhZDkzMTVlNjA5YjlmN2NiZDA4ZTFlZGQ1MTkyYWExMjdiMmMwOGFjZGVhYjZkM2MxOTBmMDUxYjNhYmFlNzI4N2E4NDIzNzgwMmZjOTg1NjNmeCQ3MGNmOTRlMy00ZjIyLTQ4ZTEtYmRlOS1jNGVmY2M2NTcyZWZ4gDk3YjQ3OTg0ZjA4YWMzOTJjMGY3MWE0MmRhYWZkNmQ3OTVkNDJlYzlkYWZjNzg2Mzk2MzAzYjgwZDM4MDQ0MGZhMjhkYTdhYzBhYjMyYjk3MTA2ZmVmZjEwMzc1NzczYzVhYWRiOGViYmM4MWNjNjYyOTVkNWU1ODZmMmQ5MTQz",
		"head": "8060272602d50c0394710a1db6bc5621d83f908682762d0d22739a65d8f2de6715fbc7f492dc01688da1f5e79c139aad7634b96d71212956a4fb998341b6e4c3",
		"link": "http://localhost:8008/state/8ed94c3c52af1d995bde32c32b4fae693f807da208dc86dcf20fa2f620e20d9edcc17d?head=8060272602d50c0394710a1db6bc5621d83f908682762d0d22739a65d8f2de6715fbc7f492dc01688da1f5e79c139aad7634b96d71212956a4fb998341b6e4c3"
	  }
	`
	var decoded struct {
		Proposals map[string]string `cbor:"proposals"`
	}
	err := client.unmarshalStatePayload(&decoded, responsePayload)

	require.NoError(t, err)
	assert.Equal(t, `63fcbc8fc94d0fded0398302fa07fa4e449a8660d6a55f61504b8c46374a16f9d281ec6e0a36d5b90f48a5454331ebf73de48240541ae434559a1949bb66ce9c`,
		decoded.Proposals["8d6d3891-50dd-418e-b3a1-5c6d2d850fc5"])
}
