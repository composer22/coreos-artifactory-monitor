// coreos-artifactory-monitor is a simple server that monitors artifactory for image version changes and
// deploys those images to a cluster.
package main

import (
	"flag"
	"runtime"
	"strings"

	"github.com/composer22/coreos-artifactory-monitor/logger"
	"github.com/composer22/coreos-artifactory-monitor/server"
)

var (
	log *logger.Logger
)

func init() {
	log = logger.New(logger.UseDefault, false)
}

// main is the main entry point for the application or server launch.
func main() {
	opts := &server.Options{}
	var showVersion bool

	flag.StringVar(&opts.Name, "N", "", "Name of the server.")
	flag.StringVar(&opts.Name, "name", "", "Name of the server.")
	flag.StringVar(&opts.HostName, "H", server.DefaultHostName, "HostName of the server.")
	flag.StringVar(&opts.HostName, "hostname", server.DefaultHostName, "HostName of the server.")
	flag.StringVar(&opts.Domain, "O", "", "Domain of the server.")
	flag.StringVar(&opts.Domain, "domain", "", "Domain of the server.")
	flag.StringVar(&opts.Environment, "E", server.DefaultEnvironment, "Environment of the cluster.")
	flag.StringVar(&opts.Environment, "environment", server.DefaultEnvironment, "Environment of the cluster.")

	flag.StringVar(&opts.DeployURL, "s", "", "URL of the coreos-deploy service.")
	flag.StringVar(&opts.DeployURL, "deploy_url", "", "URL of the coreos-deploy service.")
	flag.StringVar(&opts.DeployToken, "k", "", "Token to access the coreos-deploy service.")
	flag.StringVar(&opts.DeployToken, "deploy_token", "", "Token to access the coreos-deploy service.")

	flag.StringVar(&opts.ArtAPIEndpoint, "a", "", "Artifactory API Endpoint.")
	flag.StringVar(&opts.ArtAPIEndpoint, "art_endpoint", "", "Artifactory API Endpoint.")
	flag.StringVar(&opts.ArtUserID, "u", "", "Artifactory User ID.")
	flag.StringVar(&opts.ArtUserID, "art_userid", "", "Artifactory User ID.")
	flag.StringVar(&opts.ArtPassword, "w", "", "Artifactory Password.")
	flag.StringVar(&opts.ArtPassword, "art_password", "", "Artifactory Password.")
	flag.IntVar(&opts.ArtPollingInterval, "g", server.DefaultPollingInterval, "Artifactory polling time in seconds.")
	flag.IntVar(&opts.ArtPollingInterval, "art_polling", server.DefaultPollingInterval, "Artifactory polling time in seconds.")
	flag.StringVar(&opts.ArtDeployRepo, "t", "", "Name of the repo for deploy requests.")
	flag.StringVar(&opts.ArtDeployRepo, "art_deploy_repo", "", "Name of the repo for deploy requests.")
	flag.StringVar(&opts.ArtPayloadRepo, "y", "", "Name of the repo for payloads.")
	flag.StringVar(&opts.ArtPayloadRepo, "art_payload_repo", "", "Name of the repo for payloads.")

	flag.IntVar(&opts.Port, "p", server.DefaultPort, "Port to listen on for http requests.")
	flag.IntVar(&opts.Port, "port", server.DefaultPort, "Port to listen on for http requests.")
	flag.IntVar(&opts.ProfPort, "L", server.DefaultProfPort, "Profiler port to listen on.")
	flag.IntVar(&opts.ProfPort, "profiler_port", server.DefaultProfPort, "Profiler port to listen on.")
	flag.IntVar(&opts.MaxProcs, "X", server.DefaultMaxProcs, "Maximum processor cores to use.")
	flag.IntVar(&opts.MaxProcs, "procs", server.DefaultMaxProcs, "Maximum processor cores to use.")
	flag.StringVar(&opts.DSN, "D", "", "DSN connection string.")
	flag.StringVar(&opts.DSN, "dsn", "", "DSN connection string.")
	flag.BoolVar(&opts.Debug, "d", false, "Enable debugging output.")
	flag.BoolVar(&opts.Debug, "debug", false, "Enable debugging output.")
	flag.BoolVar(&showVersion, "V", false, "Show version.")
	flag.BoolVar(&showVersion, "version", false, "Show version.")
	flag.Usage = server.PrintUsageAndExit
	flag.Parse()

	// Version flag request?
	if showVersion {
		server.PrintVersionAndExit()
	}

	// Check additional params beyond the flags.
	for _, arg := range flag.Args() {
		switch strings.ToLower(arg) {
		case "version":
			server.PrintVersionAndExit()
		case "help":
			server.PrintUsageAndExit()
		}
	}

	// Validate the mandatory options.
	if err := opts.Validate(); err != nil {
		log.Errorf(err.Error())
		return
	}

	// Set thread and proc usage.
	if opts.MaxProcs > 0 {
		runtime.GOMAXPROCS(opts.MaxProcs)
	}
	log.Infof("NumCPU %d GOMAXPROCS: %d\n", runtime.NumCPU(), runtime.GOMAXPROCS(-1))

	s := server.New(opts, log)

	if err := s.Start(); err != nil {
		log.Errorf(err.Error())
	}
}
