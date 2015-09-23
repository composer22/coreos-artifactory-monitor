package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/composer22/coreos-artifactory-monitor/db"
	"github.com/composer22/coreos-artifactory-monitor/logger"
	coscl "github.com/composer22/coreos-deploy-client/client"
	cosddb "github.com/composer22/coreos-deploy/db"
)

// DeployWorker is a struct used to manage the deploy job to the cluster.
type DeployWorker struct {
	Name     string          `json:"name"`     // The image name to deploy.
	Version  string          `json:"version"`  // The version to deploy.
	Opts     *Options        `json:"options"`  // Server options.
	DeployID string          `json:"deployID"` // A UUID returned from the deploy.
	log      *logger.Logger  `json:"-"`        // Logger for messages.
	db       *db.DBConnect   `json:"-"`        // Database connection
	wg       *sync.WaitGroup `json:"-"`        // The wait group.
}

// NewDeployWorker is a factory function that returns a DeployWorker instance.
func NewDeployWorker(name string, version string, o *Options, l *logger.Logger,
	d *db.DBConnect, w *sync.WaitGroup) *DeployWorker {
	return &DeployWorker{
		Name:    name,
		Version: version,
		Opts:    o,
		log:     l,
		db:      d,
		wg:      w,
	}
}

// Run is a go routine that performs the deploy job actions.
func (d *DeployWorker) Run() {
	d.wg.Add(1)
	defer d.wg.Done()
	// Write the start of job record to the DB.
	d.db.StartDeploy(d.Opts.Domain, d.Opts.Environment, d.Name, d.Version)

	// Build standard file name ex: foo.com-development-video-mobile-1.0.1-23.tar.gz
	tarFilePrefix := fmt.Sprintf("%s-%s-%s-%s", d.Opts.Domain, d.Opts.Environment, d.Name, d.Version)
	tarFileName := fmt.Sprintf("%s.tar.gz", tarFilePrefix)

	tarPath := fmt.Sprintf("%s%s/", tmpDir, d.Name)          // evaluates as "/tmp/" + "Appname" => "/tmp/Appname/"
	tarFilePath := fmt.Sprintf("%s%s", tarPath, tarFileName) // evaluates as "/tmp/Appname/" + "foo.tar.gz" => "/tmp/Appname/foo.tar.gz"

	if err := os.MkdirAll(tarPath, 0744); err != nil {
		d.log.Errorf("Cannot make tar temp path %s: %s", tarPath, err.Error())
		d.db.UpdateDeployByName(d.Opts.Domain, d.Opts.Environment, d.Name, "", cosddb.Failed)
		return
	}
	// Download and untar the assets for this deploy from Artifactory.
	if errMsg := d.downloadAssets(tarPath, tarFilePath, tarFileName); errMsg != "" {
		d.log.Errorf(errMsg)
		d.db.UpdateDeployByName(d.Opts.Domain, d.Opts.Environment, d.Name, "", cosddb.Failed)
		return
	}

	defer os.Remove(tarFilePath)
	// evaluates as "/tmp/Appname/" + "foo.com-development-video-mobile-1.0.1-23" + "/"
	untarredPath := fmt.Sprintf("%s%s/", tarPath, tarFilePrefix)
	defer os.RemoveAll(untarredPath)

	// Get the filenames from the directory:
	metaFileName, serviceFileName, etcd2FileName := "", "", ""
	files, _ := ioutil.ReadDir(untarredPath)
	for _, f := range files {
		name := path.Base(f.Name())
		switch path.Ext(name) {
		case ".json":
			metaFileName = name
		case ".service":
			serviceFileName = name
		case ".tmpl":
			serviceFileName = name
		case ".etcd2":
			etcd2FileName = name
		default:
		}
	}

	// Validate deploy files exist and set paths.
	if metaFileName == "" {
		d.log.Errorf("Metadata file not found in %s", tarFilePath)
		d.db.UpdateDeployByName(d.Opts.Domain, d.Opts.Environment, d.Name, "", cosddb.Failed)
		return
	}
	if serviceFileName == "" {
		d.log.Errorf("Service unit file not found in %s", tarFilePath)
		d.db.UpdateDeployByName(d.Opts.Domain, d.Opts.Environment, d.Name, "", cosddb.Failed)
		return
	}
	metaFilePath := fmt.Sprintf("%s%s", untarredPath, metaFileName)
	serviceFilePath := fmt.Sprintf("%s%s", untarredPath, serviceFileName)
	etcd2FilePath := ""
	if etcd2FileName != "" {
		etcd2FilePath = fmt.Sprintf("%s%s", untarredPath, etcd2FileName)
	}

	// Get the metadata from the file.
	metaData, errMsg := d.getMetaData(metaFilePath)
	if errMsg != "" {
		d.log.Errorf(errMsg)
		d.db.UpdateDeployByName(d.Opts.Domain, d.Opts.Environment, d.Name, "", cosddb.Failed)
		return
	}

	// Submit a deploy request to the client library.
	co := &coscl.Options{
		Name:             metaData.Name,
		Version:          metaData.Version,
		ImageVersion:     metaData.ImageVersion,
		NumInstances:     metaData.NumInstances,
		TemplateFilePath: serviceFilePath,
		Etcd2FilePath:    etcd2FilePath,
		Token:            d.Opts.DeployToken,
		Url:              d.Opts.DeployURL,
		Debug:            false,
	}

	cl := coscl.New(co) // API client
	deployID, errMsg := d.submitDeployRequest(cl)
	if errMsg != "" {
		d.log.Errorf(errMsg)
		d.db.UpdateDeployByName(d.Opts.Domain, d.Opts.Environment, d.Name, "", cosddb.Failed)
		return
	}
	co.DeployID = deployID
	d.DeployID = deployID

	// Loop check the status of the deploy and wait for the deploy to complete. Timeout 1 minute.
	errMsg = d.submitStatusRequest(cl, deployID)
	if errMsg != "" {
		d.log.Errorf(errMsg)
		d.db.UpdateDeployByName(d.Opts.Domain, d.Opts.Environment, d.Name, deployID, cosddb.Failed)
		return
	}

	// Mark the job complete.
	d.db.UpdateDeployByName(d.Opts.Domain, d.Opts.Environment, d.Name, d.DeployID, cosddb.Success)
}

// downloadAssets retrieves and untars the assets from the Artifactory repository.
func (d *DeployWorker) downloadAssets(tarPath string, tarFilePath string, tarFileName string) string {
	artFilePath := strings.Replace(d.Opts.ArtAPIEndpoint, "/api", "", 1) // No API.
	httpPath := fmt.Sprintf("%s/%s/%s/%s", artFilePath, d.Opts.ArtPayloadRepo, d.Name, tarFileName)
	req, err := http.NewRequest(httpGet, httpPath, nil)
	if err != nil {
		return fmt.Sprintf("Cannot create request for %s: %s", httpPath, err.Error())
	}
	req.SetBasicAuth(d.Opts.ArtUserID, d.Opts.ArtPassword)
	cl := &http.Client{}
	resp, err := cl.Do(req)
	if err != nil {
		return fmt.Sprintf("Cannot retrieve file for %s: %s", httpPath, err.Error())
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return fmt.Sprintf("Cannot read body for file %s: %s", httpPath, err.Error())
	}
	err = ioutil.WriteFile(tarFilePath, body, 0644)
	if err != nil {
		return fmt.Sprintf("Cannot write file %s: %s", tarFilePath, err.Error())
	}

	// Untar the assets.
	cmd := exec.Command("tar", "-xzf", tarFilePath, "-C", tarPath)
	if _, err := execCmd(cmd); err != nil {
		return fmt.Sprintf("Cannot untar file %s: %s", tarFilePath, err.Error())
	}
	return ""
}

// getMetaData returns the metadata from the untarred file for the application assets.
func (d *DeployWorker) getMetaData(metaFilePath string) (*coscl.ServiceTemplateVars, string) {
	var metaData coscl.ServiceTemplateVars
	m, err := ioutil.ReadFile(metaFilePath)
	if err != nil {
		return nil, fmt.Sprintf("Cannot read metadata from %s: %s", metaFilePath, err.Error())
	}
	err = json.Unmarshal(m, &metaData)
	if err != nil {
		return nil, fmt.Sprintf("Cannot parse metadata from %s: %s", metaFilePath, err.Error())
	}
	return &metaData, ""
}

// submitDeployRequest returns a unique deploy id after submitting a request via the client library to
// the coreos-deploy service in the cluster.
func (d *DeployWorker) submitDeployRequest(cl *coscl.Client) (string, string) {
	resp, err := cl.Execute()
	if err != nil {
		return "", fmt.Sprintf("Could not submit deploy request: %s", err.Error())
	}

	result := struct {
		DeployID string `json:"deployID"` // The UUID of the deploy.
	}{}
	err = json.Unmarshal([]byte(resp), &result)
	if err != nil {
		return "", fmt.Sprintf("Cannot parse returned deploy id: %s", err.Error())
	}
	return result.DeployID, ""
}

// submitStatusRequest checks the service to validate that the deploy request completed successfully.
func (d *DeployWorker) submitStatusRequest(cl *coscl.Client, deployID string) string {
	var stat cosddb.DeployStatus

	// Check status every maxPollStatusPause seconds, maxPollStatusCount times.
	for a := 0; a <= maxPollStatusCount; a++ {
		time.Sleep(time.Second * maxPollStatusPause)
		resp, err := cl.Execute()
		if err != nil {
			return fmt.Sprintf("Could not submit status request for deployID %s", deployID, err.Error())
		}
		err = json.Unmarshal([]byte(resp), &stat)
		if err != nil {
			return fmt.Sprintf("Could not parse status response for deployID %s: %s", deployID, err.Error())
		}
		if stat.Status == cosddb.Started {
			continue
		}
		break
	}
	switch stat.Status {
	case cosddb.Started:
		return fmt.Sprintf("Deploy still running after polling period expired for deployID %s", deployID)
	case cosddb.Failed:
		return fmt.Sprintf("Deploy Failed for deployID %s", deployID)
	default:
	}
	return ""
}
