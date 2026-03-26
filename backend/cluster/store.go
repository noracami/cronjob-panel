// backend/cluster/store.go
package cluster

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Cluster struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Endpoint  string    `json:"endpoint"`
	AuthType  string    `json:"auth_type"`
	AuthData  []byte    `json:"-"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

func CreateCluster(db *sql.DB, c Cluster) error {
	c.ID = uuid.New().String()
	_, err := db.Exec(
		"INSERT INTO clusters (id, name, endpoint, auth_type, auth_data, created_by) VALUES (?, ?, ?, ?, ?, ?)",
		c.ID, c.Name, c.Endpoint, c.AuthType, c.AuthData, c.CreatedBy,
	)
	return err
}

func ListClusters(db *sql.DB) ([]Cluster, error) {
	rows, err := db.Query("SELECT id, name, endpoint, auth_type, created_by, created_at FROM clusters")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clusters []Cluster
	for rows.Next() {
		var c Cluster
		if err := rows.Scan(&c.ID, &c.Name, &c.Endpoint, &c.AuthType, &c.CreatedBy, &c.CreatedAt); err != nil {
			return nil, err
		}
		clusters = append(clusters, c)
	}
	return clusters, nil
}

func GetCluster(db *sql.DB, id string) (*Cluster, error) {
	var c Cluster
	err := db.QueryRow(
		"SELECT id, name, endpoint, auth_type, auth_data, created_by, created_at FROM clusters WHERE id = ?", id,
	).Scan(&c.ID, &c.Name, &c.Endpoint, &c.AuthType, &c.AuthData, &c.CreatedBy, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func DeleteCluster(db *sql.DB, id string) error {
	_, err := db.Exec("DELETE FROM clusters WHERE id = ?", id)
	return err
}
