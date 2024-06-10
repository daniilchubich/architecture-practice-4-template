package integration

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"
)

const baseAddress = "http://balancer:8090"
const team = "gods"

var client = http.Client{
	Timeout: 3 * time.Second,
}

func getData(key string) (*http.Response, error) {
	path := fmt.Sprintf("%s/api/v1/some-data", baseAddress)

	queryParams := url.Values{}
	queryParams.Set("key", key)
	path += "?" + queryParams.Encode()

	return client.Get(path)
}

func TestBalancer(t *testing.T) {
	if _, exists := os.LookupEnv("INTEGRATION_TEST"); !exists {
		t.Skip("Integration test is not enabled")
	}


	senders := make(map[string]bool)
	for i := 0; i < 10; i++ {
		resp, err := getData(team)
	if err != nil {
		t.Error(err)
	}
	sender := resp.Header.Get("lb-from")
	t.Logf("response from [%s]", sender)
	senders[sender] = true
	}

	count := 0
	for range senders {
		count++
	}

	if count < 3 {
		t.Errorf("expected at least 3 senders, got %d", count)
	}

	resp, err := getData(team)
	if err != nil {
		t.Error(err)
	}
	if (resp.StatusCode != http.StatusOK) {
		t.Errorf("expected status code 200, got %d", resp.StatusCode)
	}

	resp, err = getData("bad-key")
	if err != nil {
		t.Error(err)
	}
	if (resp.StatusCode != http.StatusNotFound) {
		t.Errorf("expected status code 404, got %d", resp.StatusCode)
	}
}

func BenchmarkBalancer(b *testing.B) {
    if _, exists := os.LookupEnv("INTEGRATION_TEST"); !exists {
        b.Skip("Integration test is not enabled")
    }

    senders := make(map[string]int)
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        resp, err := getData(team)
        if err != nil {
            b.Error(err)
        }
        sender := resp.Header.Get("lb-from")
        senders[sender]++
        resp.Body.Close()
    }

    b.StopTimer()

    count := 0
    for _, v := range senders {
        if v > 0 {
            count++
        }
    }

    b.Logf("Requests were served by %d senders", count)
}