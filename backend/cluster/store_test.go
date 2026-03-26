// backend/cluster/store_test.go
package cluster

import (
	"database/sql"
	"os"
	"testing"

	"github.com/kerke/cronjob-panel/db"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	path := t.TempDir() + "/test.db"
	t.Cleanup(func() { os.Remove(path) })

	database, err := db.InitDB(path)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	// Insert a test user to satisfy FK constraints
	_, err = database.Exec(
		"INSERT INTO users (id, discord_id, username) VALUES (?, ?, ?)",
		"test-user-001", "discord-001", "testuser",
	)
	if err != nil {
		t.Fatalf("insert test user failed: %v", err)
	}

	return database
}

func TestCreateAndListClusters(t *testing.T) {
	database := setupTestDB(t)

	c := Cluster{
		Name:      "test-cluster",
		Endpoint:  "https://k8s.example.com",
		AuthType:  "token",
		AuthData:  []byte("secret-token"),
		CreatedBy: "test-user-001",
	}

	if err := CreateCluster(database, c); err != nil {
		t.Fatalf("CreateCluster failed: %v", err)
	}

	clusters, err := ListClusters(database)
	if err != nil {
		t.Fatalf("ListClusters failed: %v", err)
	}

	if len(clusters) != 1 {
		t.Fatalf("expected 1 cluster, got %d", len(clusters))
	}

	if clusters[0].Name != "test-cluster" {
		t.Errorf("expected name %q, got %q", "test-cluster", clusters[0].Name)
	}
}

func TestDeleteCluster(t *testing.T) {
	database := setupTestDB(t)

	c := Cluster{
		Name:      "to-delete",
		Endpoint:  "https://k8s.example.com",
		AuthType:  "token",
		AuthData:  []byte("secret-token"),
		CreatedBy: "test-user-001",
	}

	if err := CreateCluster(database, c); err != nil {
		t.Fatalf("CreateCluster failed: %v", err)
	}

	clusters, err := ListClusters(database)
	if err != nil {
		t.Fatalf("ListClusters failed: %v", err)
	}
	if len(clusters) != 1 {
		t.Fatalf("expected 1 cluster before delete, got %d", len(clusters))
	}

	if err := DeleteCluster(database, clusters[0].ID); err != nil {
		t.Fatalf("DeleteCluster failed: %v", err)
	}

	remaining, err := ListClusters(database)
	if err != nil {
		t.Fatalf("ListClusters after delete failed: %v", err)
	}

	if len(remaining) != 0 {
		t.Errorf("expected 0 clusters after delete, got %d", len(remaining))
	}
}
