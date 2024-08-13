package adguard

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestListRecords(t *testing.T) {
	tests := []struct {
		name            string
		serverResponses []serverResponse
		expectedResult  map[string][]string
		expectedError   bool
	}{
		{
			name: "Successful API call",
			serverResponses: []serverResponse{
				{statusCode: http.StatusOK, body: `[{"Name":"record1","Content":"content1"},{"Name":"record2","Content":"content1"},{"Name":"record3","Content":"content2"}]`},
			},
			expectedResult: map[string][]string{
				"content1": {"record1", "record2"},
				"content2": {"record3"},
			},
			expectedError: false,
		},
		{
			name: "Retry on non-200 status code",
			serverResponses: []serverResponse{
				{statusCode: http.StatusInternalServerError, body: ""},
				{statusCode: http.StatusOK, body: `[{"Name":"record1","Content":"content1"}]`},
			},
			expectedResult: map[string][]string{
				"content1": {"record1"},
			},
			expectedError: false,
		},
		{
			name: "Max retries exceeded",
			serverResponses: []serverResponse{
				{statusCode: http.StatusInternalServerError, body: ""},
				{statusCode: http.StatusInternalServerError, body: ""},
				{statusCode: http.StatusInternalServerError, body: ""},
			},
			expectedResult: nil,
			expectedError:  true,
		},
		{
			name: "Invalid JSON response",
			serverResponses: []serverResponse{
				{statusCode: http.StatusOK, body: `invalid json`},
			},
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "Bearer testToken", r.Header.Get("Authorization"))

				response := tt.serverResponses[0]
				tt.serverResponses = tt.serverResponses[1:]
				w.WriteHeader(response.statusCode)
				w.Write([]byte(response.body))
			}))
			defer server.Close()

			client := &Client{
				Url:        server.URL,
				authHeader: "Bearer testToken",
				Retries:    2,
				RetryDelay: time.Millisecond,
			}

			result, err := client.ListRecords()

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

type serverResponse struct {
	statusCode int
	body       string
}
