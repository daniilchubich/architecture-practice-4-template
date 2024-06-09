package integration

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

const baseAddress = "http://balancer:8090"

var client = http.Client{
	Timeout: 3 * time.Second,
}

func TestBalancer(t *testing.T) {
	if _, exists := os.LookupEnv("INTEGRATION_TEST"); !exists {
		t.Skip("Integration test is not enabled")
	}


	senders := make(map[string]bool)
	for i := 0; i < 10; i++ {
		resp, err := client.Get(fmt.Sprintf("%s/api/v1/some-data", baseAddress))
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
}

func BenchmarkBalancer(b *testing.B) {
    if _, exists := os.LookupEnv("INTEGRATION_TEST"); !exists {
        b.Skip("Integration test is not enabled")
    }

    senders := make(map[string]int)
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        resp, err := client.Get(fmt.Sprintf("%s/api/v1/some-data", baseAddress))
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