package db

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

const (
	_ = iota
	Started
	Success
	Failed
)

type DBConnect struct {
	db *sql.DB
}

// NewDBConnect is a factory method that returns a new db connection
func NewDBConnect(dsn string) (*DBConnect, error) {

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Validate DSN data
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &DBConnect{db: db}, nil
}

// ValidAuth returns true if the API Key is valid for a request.
func (d *DBConnect) ValidAuth(key string) bool {
	var id int
	row := d.db.QueryRow("SELECT id FROM artifactory_auth_tokens WHERE token = ?", key)
	err := row.Scan(&id)

	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		return false
	default:
		return true
	}
}

// StartDeploy inserts or updates the deploy tracker for versions.
func (d *DBConnect) StartDeploy(domain string, environment string, name string, version string) bool {
	// Compound unique key: domain, environment, name
	result, err := d.db.Exec("INSERT INTO artifactory_deploys (domain, environment, service_name, version, "+
		"status, updated_at, created_at) "+
		"VALUES (?, ?, ?, ?, ?,  NOW(), NOW())"+
		"ON DUPLICATE KEY UPDATE status = ?, version = ?, updated_at = NOW()",
		domain, environment, name, version, Started, Started, version)
	if err != nil {
		return false
	}
	id, err := result.LastInsertId()
	if err != nil || id <= 0 {
		return false
	}
	return true
}

// UpdateDeploy updates the deploy row with information from the run.
func (d *DBConnect) UpdateDeployByName(domain string, environment string, name string,
	deployID string, status int) bool {
	result, err := d.db.Exec("UPDATE artifactory_deploys "+
		"SET deploy_id = ?, "+
		"status = ?, "+
		"updated_at = NOW() "+
		"WHERE domain = ?, environment = ?, service_name = ?",
		deployID, status, domain, environment, name)
	if err != nil {
		return false
	}
	rows, err := result.RowsAffected()
	if err != nil || rows != 1 {
		return false
	}
	return true
}

// DeployStatus is used to return deploy status information from the database to the requester.
type DeployStatus struct {
	DeployID    string `json:"deployID"`    // The deploy UUID.
	Domain      string `json:"domain"`      // The domain name serviced.
	Environment string `json:"environment"` // The environment serviced (development, qa etc.)
	Name        string `json:"name"`        // The application name of the service ex: video-mobile.
	Version     string `json:"version"`     // The version of the application ex; 1.0.0-32
	Status      int    `json:"status"`      // The status ID of the result.
	UpdatedAt   string `json:"updatedAt"`   // The create date and time of the deploy.
	CreatedAt   string `json:"createdAt"`   // The last update to this record.
}

// QueryDeploy returns the status of a deploy request.
func (d *DBConnect) QueryDeployByName(domain string, environment string, name string) (*DeployStatus, error) {
	r := &DeployStatus{}
	row := d.db.QueryRow("SELECT deploy_id, domain, environment, service_name, version, status, updated_at, created_at "+
		"FROM artifactory_deploys WHERE service_name = ?", name)
	err := row.Scan(&r.DeployID, &r.Domain, &r.Environment, &r.Name, &r.Version, &r.Status, &r.UpdatedAt, &r.CreatedAt)
	switch {
	case err == sql.ErrNoRows:
		return nil, err
	case err != nil:
		return nil, err
	default:
		return r, nil
	}
}

// Close closes the connection(s) to the DB.
func (d *DBConnect) Close() bool {
	d.db.Close()
	return true
}
