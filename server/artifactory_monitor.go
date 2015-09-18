package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	cosddb "github.com/composer22/coreos-deploy/db"
)

// Monitor is a go routine that continually monitors artifactory for any version changes.
func (s *Server) Monitor() {
	s.wg.Add(1)
	defer s.wg.Done()

	var wg sync.WaitGroup
	for {
		timer := time.NewTimer(time.Second * time.Duration(s.opts.ArtPollingInterval))
		select {
		case <-s.done: // Shutdown signal.
			timer.Stop()
			return
		case <-s.force: // Force a check for deltas. Don't wait.
			timer.Stop()
		case <-timer.C: // Timeout.
		}

		// Get changes.
		deploys, err := s.checkDeltas(&wg)
		if err != nil {
			s.log.Errorf("Check Deltas Error: %s", err.Error())
		}

		// run the deploys.
		for _, d := range deploys {
			go d.Run()
		}
		wg.Wait() // Wait for all deploy jobs to complete before monitoring again.
	}
}

// ArtFolderInfo is returned from a call to collect folder info from an API request.
type ArtFolderInfo struct {
	Repo         string                `json:"repo"`         // The repository being queried.
	Path         string                `json:"path"`         // The file path in the repo.
	Created      string                `json:"created"`      // When the folder or file was created.
	CreatedBy    string                `json:"createdBy"`    // Who created the folder or file.
	LastModified string                `json:"lastModified"` // When was the folder or file last modified.
	ModifiedBy   string                `json:"modifiedBy"`   // Who modified the folder or file.
	LastUpdated  string                `json:"lastUpdated"`  // This might be the same as last modified?
	Children     []*ArtFolderInfoChild `json:"children"`     // This is a list of folders and files in the directory.
	Uri          string                `json:"uri"`          // The API URL that was called.
}

// ArtFolderInfoChild is returned from a call to collect folder info from an API request. See ArtFolderInfo.
type ArtFolderInfoChild struct {
	Uri    string `json:"uri"`    // The subdirectory or file name ex: "/1.0.0-21"
	Folder bool   `json:"folder"` // If true, this is a subdirectory.
}

// checkDeltas returns an array of deploy jobs, one for each docker instance who's version has changed in artifactory.
func (s *Server) checkDeltas(wg *sync.WaitGroup) ([]*DeployWorker, error) {
	jobs := make([]*DeployWorker, 0)

	// Get image names from repo.
	images, err := s.getArtFolders(s.opts.ArtImageRepo)
	if err != nil {
		return nil, err
	}

	// Check each image for the latest version and add it to the deploy list if needed.
	for _, image := range images {
		// dir equates as "reponame" + "/docker-image-version" => "foorepo/1.0.0-23"
		dir := fmt.Sprintf("%s%s", s.opts.ArtImageRepo, image.Uri)
		versions, err := s.getArtFolders(dir)
		if err != nil {
			s.log.Errorf("Unable to read directory %s: %s", dir, err.Error())
			continue
		}
		// Find the latest version tag for this image.
		var latestVersion string = ""
		for _, iv := range versions {
			if iv.Uri > latestVersion {
				latestVersion = iv.Uri
			}
		}
		imageName := strings.Replace(image.Uri, "/", "", 1)
		latestVersion = strings.Replace(latestVersion, "/", "", 1)

		// Check the last version deployed from the database.
		lastDep, err := s.db.QueryDeployByName(s.opts.Domain, s.opts.Environment, imageName)
		if err != nil && err != sql.ErrNoRows {
			s.log.Errorf("Unable to read deploy from db for %s-%s-%s: %s", s.opts.Domain,
				s.opts.Environment, imageName, err.Error())
			continue
		}
		// If no version has been deployed, or it's out of date, or it failed before
		// then create a new job.
		if err != nil || lastDep.Version < latestVersion ||
			(lastDep.Version == latestVersion && lastDep.Status == cosddb.Failed) {
			jobs = append(jobs, NewDeployWorker(imageName, latestVersion, s.opts, s.log, s.db, wg))
		}
	}
	return jobs, nil
}

// getArtFolders retrieves a list of folders from the Artifactory directory path.
func (s *Server) getArtFolders(subdir string) ([]*ArtFolderInfoChild, error) {
	results := make([]*ArtFolderInfoChild, 0)
	// evaluates as "http://art.com/foo/api" + "/storage" + "/" + "sub/directory"
	req, err := http.NewRequest(httpGet, fmt.Sprintf("%s%s/%s/", s.opts.ArtAPIEndpoint, artSourceRoute, subdir), nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(s.opts.ArtUserID, s.opts.ArtPassword)
	resp, err := s.sendRequest(req)
	if err != nil {
		return nil, err
	}
	var fi ArtFolderInfo
	err = json.Unmarshal([]byte(resp), fi)
	if err != nil {
		return nil, err
	}

	// Filter out files or folders named "latest"
	for _, c := range fi.Children {
		if c.Folder == false || c.Uri == "/latest" {
			continue
		}
		results = append(results, c)
	}
	return results, nil
}

// sendRequest sends a request to a server and prints the result.
func (s *Server) sendRequest(req *http.Request) (string, error) {
	cl := &http.Client{}
	resp, err := cl.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
