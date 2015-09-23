package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
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
			//		case <-timer.C: // Timeout.
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

	// Get folders names from repo.
	apps, err := s.getArtFolders(s.opts.ArtDeployRepo, true)
	if err != nil {
		return nil, err
	}

	// Check each folder find the latest deploy version and add it to the deploy list if needed.
	for _, app := range apps {
		// dir equates as "reponame" + "/appname" => "foorepo/appname"
		dir := fmt.Sprintf("%s%s", s.opts.ArtDeployRepo, app.Uri)
		versions, err := s.getArtFolders(dir, false)
		if err != nil {
			s.log.Errorf("Unable to read directory %s: %s", dir, err.Error())
			continue
		}
		// Find the latest version tag for this application.
		var latestVersion string = ""
		for _, iv := range versions {
			if iv.Uri > latestVersion {
				latestVersion = iv.Uri
			}
		}
		appName := strings.Replace(app.Uri, "/", "", 1)
		latestVersion = strings.Replace(latestVersion, "/", "", 1)
		latestVersion = strings.Replace(latestVersion, ".deploy", "", 1)

		// Check the last version deployed from the database.
		lastDep, err := s.db.QueryDeployByName(s.opts.Domain, s.opts.Environment, appName)
		if err != nil && err != sql.ErrNoRows {
			s.log.Errorf("Unable to read deploy from db for %s-%s-%s: %s", s.opts.Domain,
				s.opts.Environment, appName, err.Error())
			continue
		}
		// If no version has been deployed, or it's out of date, or it failed before then create a new job.
		if err != nil || lastDep.Version < latestVersion ||
			(lastDep.Version == latestVersion && lastDep.Status == cosddb.Failed) {
			jobs = append(jobs, NewDeployWorker(appName, latestVersion, s.opts, s.log, s.db, wg))
		}
	}
	return jobs, nil
}

// getArtFolders retrieves a list from the Artifactory directory path contents.
func (s *Server) getArtFolders(subdir string, retrieveFolders bool) ([]*ArtFolderInfoChild, error) {
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
	err = json.Unmarshal([]byte(resp), &fi)
	if err != nil {
		return nil, err
	}

	// Retrieve content of subdir.
	for _, c := range fi.Children {
		// Filter out files vs folders.
		if c.Folder != retrieveFolders {
			continue
		}
		// If file check and not a deploy file, continue scan.
		if !retrieveFolders && path.Ext(c.Uri) != ".deploy" {
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
