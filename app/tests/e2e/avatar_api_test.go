//go:build e2e

package e2e_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iPatrushevSergey/gophprofile/app/tests/testsupport"
)

func TestE2E_Avatar_uploadAndMetadata(t *testing.T) {
	srv := testsupport.NewTestServer(t)
	base := srv.APIBase()
	client := srv.Server.Client()
	userID := testsupport.E2EUserID

	resp := testsupport.UploadAvatar(t, client, base, userID, "avatar.png", "image/png", testsupport.MinimalPNG(t))
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	uploaded := testsupport.DecodeUploadResponse(t, resp)
	require.NotEmpty(t, uploaded.ID)

	meta := testsupport.GetMetadataViaAPI(t, client, base, uploaded.ID)
	assert.Equal(t, uploaded.ID, meta.ID)
	assert.Equal(t, userID, meta.UserID)
	assert.Equal(t, "avatar.png", meta.FileName)
}

func TestE2E_Avatar_uploadRequiresUserHeader(t *testing.T) {
	srv := testsupport.NewTestServer(t)
	client := srv.Server.Client()

	req, err := http.NewRequest(http.MethodPost, srv.APIBase()+"/api/v1/avatars", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
