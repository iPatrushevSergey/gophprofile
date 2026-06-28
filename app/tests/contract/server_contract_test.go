//go:build contract

package contract_test

import (
	"encoding/json"
	"net/http"
	"testing"

	easyjson "github.com/mailru/easyjson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	processormodel "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/adapters/broker/rabbitmq/model"
	servermodel "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/rabbitmq/model"
	presdto "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/presentation/http/dto"
	"github.com/iPatrushevSergey/gophprofile/app/tests/testsupport"
)

// Producer contract: server health endpoint exposes the agreed JSON shape.
func TestServerHealth_producerContract(t *testing.T) {
	srv := testsupport.NewTestServer(t)

	resp, err := srv.Server.Client().Get(srv.APIBase() + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var out presdto.HealthResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	assert.NotEmpty(t, out.Status)
	assert.NotEmpty(t, out.Database)
	assert.NotEmpty(t, out.Storage)
	assert.NotEmpty(t, out.Broker)
}

// Consumer contract: processor must decode avatar uploaded events published by the server.
func TestProcessorConsumer_avatarUploadedWireContract(t *testing.T) {
	wire := servermodel.AvatarUploadedEvent{
		AvatarID: "avatar-1",
		UserID:   "user-1",
		S3Key:    "user-1/avatar-1/original",
	}

	payload, err := easyjson.Marshal(wire)
	require.NoError(t, err)

	var consumed processormodel.AvatarUploadedEvent
	require.NoError(t, easyjson.Unmarshal(payload, &consumed))
	assert.Equal(t, wire.AvatarID, consumed.AvatarID)
	assert.Equal(t, wire.UserID, consumed.UserID)
	assert.Equal(t, wire.S3Key, consumed.S3Key)
}
