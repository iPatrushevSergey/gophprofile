//go:build component

package component_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iPatrushevSergey/gophprofile/app/tests/testsupport"
)

// Server avatar component: HTTP → avatar use cases → Postgres + MinIO + RabbitMQ.
func TestServerAvatarComponent_uploadAndList(t *testing.T) {
	srv := testsupport.NewTestServer(t)
	base := srv.APIBase()
	client := srv.Server.Client()
	userID := "component-user"

	resp := testsupport.UploadAvatar(t, client, base, userID, "avatar.png", "image/png", testsupport.MinimalPNG(t))
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	uploaded := testsupport.DecodeUploadResponse(t, resp)
	require.NotEmpty(t, uploaded.ID)
	assert.Equal(t, userID, uploaded.UserID)

	list := testsupport.ListUserAvatarsViaAPI(t, client, base, userID)
	require.Len(t, list.Items, 1)
	assert.Equal(t, uploaded.ID, list.Items[0].ID)
	assert.Equal(t, "avatar.png", list.Items[0].FileName)
}
