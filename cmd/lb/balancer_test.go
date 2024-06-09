package main

import (
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
    testCases := []struct {
        name     string
        input    string
        expected uint32
    }{
        {
            name:     "Test case 1",
            input:    "192.168.0.1",
            expected: 111728439,
        },
        {
            name:     "Test case 2",
            input:    "10.0.0.1",
            expected: 1148388597,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result := hash(tc.input)
            assert.Equal(t, tc.expected, result)
        })
    }
}

func TestHealth(t *testing.T) {
	domain := "test.com"
	mockURL := "http://test.com/health"
	
	httpmock.RegisterResponder(http.MethodGet, mockURL, httpmock.NewStringResponder(http.StatusOK, ""))
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	result := health(domain)
	assert.True(t, result)

	httpmock.RegisterResponder(http.MethodGet, mockURL, httpmock.NewStringResponder(http.StatusInternalServerError, ""))
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	result2 := health(domain)
	assert.False(t, result2)
}
