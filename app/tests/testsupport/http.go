//go:build e2e || component || contract

package testsupport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/require"

	authmw "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/presentation/http/middleware/auth"
	presdto "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/presentation/http/dto"
)

// DoJSON sends a JSON request and returns the response.
func DoJSON(t *testing.T, client *http.Client, method, url string, body any) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, url, reader)
	require.NoError(t, err)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}

// MinimalPNG returns a tiny valid PNG image for upload tests.
func MinimalPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))
	return buf.Bytes()
}

// UploadAvatar posts multipart upload to the protected avatars endpoint.
func UploadAvatar(t *testing.T, client *http.Client, baseURL, userID, fileName, mimeType string, content []byte) *http.Response {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, fileName))
	header.Set("Content-Type", mimeType)
	part, err := writer.CreatePart(header)
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req, err := http.NewRequest(http.MethodPost, baseURL+"/api/v1/avatars", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set(authmw.UserIDHeader, userID)

	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}

// DecodeUploadResponse decodes upload JSON body.
func DecodeUploadResponse(t *testing.T, resp *http.Response) presdto.UploadAvatarResponse {
	t.Helper()
	defer resp.Body.Close()
	var out presdto.UploadAvatarResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	return out
}

// GetMetadataViaAPI fetches avatar metadata by id.
func GetMetadataViaAPI(t *testing.T, client *http.Client, baseURL, avatarID string) presdto.AvatarMetadataResponse {
	t.Helper()
	resp, err := client.Get(baseURL + "/api/v1/avatars/" + avatarID + "/metadata")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var out presdto.AvatarMetadataResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	return out
}

// ListUserAvatarsViaAPI returns avatars for a user.
func ListUserAvatarsViaAPI(t *testing.T, client *http.Client, baseURL, userID string) presdto.UserAvatarsResponse {
	t.Helper()
	resp, err := client.Get(baseURL + "/api/v1/users/" + userID + "/avatars")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var out presdto.UserAvatarsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	return out
}
