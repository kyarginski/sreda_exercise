package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"sreda/internal/lib/logger/sl"

	"github.com/stretchr/testify/assert"
)

func TestProcessRequest1(t *testing.T) {
	// env := "local"
	env := "nop"
	log := sl.SetupLogger(env)

	tests := []struct {
		name string
		body string
		want int
	}{
		{
			name: "Good request",
			body: `{"iteration": 123}`,
			want: http.StatusOK,
		},
		{
			name: "Empty body request",
			body: ``,
			want: http.StatusBadRequest,
		},
		{
			name: "I am teapot request",
			body: `{"iteration777": 42}`,
			want: http.StatusTeapot,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			requestBody := tt.body
			req, err := http.NewRequest("POST", "/test", bytes.NewBufferString(requestBody))
			assert.NoError(t, err)
			w := httptest.NewRecorder()

			handler := ProcessRequest(log)
			handler(w, req)

			assert.Equal(t, tt.want, w.Code, "Status codes are equal")
		})
	}
}
