// backend/cluster/client_test.go
package cluster

import (
	"testing"
)

func TestNewK8sClient_InvalidToken(t *testing.T) {
	client, err := NewK8sClient("https://127.0.0.1:99999", "token", []byte("fake-token"))
	if err != nil {
		t.Fatalf("NewK8sClient: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}
