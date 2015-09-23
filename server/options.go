package server

import (
	"encoding/json"
	"errors"
)

// Options represents parameters that are passed to the application to be used in constructing
// the server.
type Options struct {
	Name               string `json:"name"`               // The name of the server.
	HostName           string `json:"hostName"`           // The hostname of the server.
	Domain             string `json:"domain"`             // The domain of the server.
	Environment        string `json:"environment"`        // The environment of the server (dev, stage, prod, etc).
	DeployURL          string `json:"deployURL"`          // The coreos-deploy url endpoint.
	DeployToken        string `json:"-"`                  // The coreos-deploy token for security access.
	ArtAPIEndpoint     string `json:"artAPIEndpoint"`     // The artifactory API endpoint.
	ArtUserID          string `json:"-"`                  // The artifactory user id.
	ArtPassword        string `json:"-"`                  // The artifactory password.
	ArtPollingInterval int    `json:"artPollingInterval"` // The artifactory polling interval in seconds.
	ArtDeployRepo      string `json:"artDeployRepo"`      // The artifactory repo of the deploy request files.
	ArtPayloadRepo     string `json:"artPayloadRepo"`     // The artifactory repo of the deployment payloads.
	Port               int    `json:"port"`               // The default port of the server.
	ProfPort           int    `json:"profPort"`           // The profiler port of the server.
	DSN                string `json:"-"`                  // The DSN login string to the database.
	MaxProcs           int    `json:"maxProcs"`           // The maximum number of processor cores available.
	Debug              bool   `json:"debugEnabled"`       // Is debugging enabled in the application or server.
}

// Validate options
// TBD: Fix these validations for the current keyset
func (o *Options) Validate() error {
	if o.Domain == "" {
		return errors.New("Service domain is mandatory.")
	}
	if o.DeployURL == "" {
		return errors.New("Service deploy URL is mandatory.")
	}
	if o.DeployToken == "" {
		return errors.New("Service deploy authorization token is mandatory.")
	}
	if o.ArtAPIEndpoint == "" {
		return errors.New("Artifactory API endpoint is mandatory.")
	}
	if o.ArtUserID == "" {
		return errors.New("Artifactory API user id is mandatory.")
	}
	if o.ArtDeployRepo == "" {
		return errors.New("Artifactory API deploy request repo name is mandatory.")
	}
	if o.ArtPayloadRepo == "" {
		return errors.New("Artifactory API payload repo is mandatory.")
	}
	if o.DSN == "" {
		return errors.New("DNS database settings are mandatory.")
	}
	return nil
}

// String is an implentation of the Stringer interface so the structure is returned as a string
// to fmt.Print() etc.
func (o *Options) String() string {
	b, _ := json.Marshal(o)
	return string(b)
}
