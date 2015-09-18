package server

import (
	"fmt"
	"os"
)

const usageText = `
Description: coreos-artifactory-monitor is a server for monitoring deploy needs from Artifactory to a coreos cluster.

Usage: coreos-artifactory-monitor [options...]

Server options:
    -N, --name NAME                  NAME of the server (default: empty field).
    -H, --hostname HOSTNAME          HOSTNAME of the server (default: localhost).
    -O, --domain DOMAIN              DOMAIN of the site being managed (default: localhost).
    -E, --environment ENVIRONMENT    ENVIRONMENT (development, qa, staging, production).
    -s, --deploy_url URL             URL to the coreos-deploy service.
    -k, --deploy_token TOKEN         Security TOKEN to access the coreos-deploy service.
    -a, --artif_endpoint APIURL      The base APIURL to the artifactory API service.
    -u, --artif_userid USERID        USERID to login to the artifactory API service.
    -w, --artif_password PASSWORD    PASSWORD to login to the artifactory API service.
    -g, --artif_polling INTERVAL     How often to check artifactory for deploys in INTERVAL seconds (default: 300 sec).
    -t, --artif_image_repo REPO      The name of the REPO where Docker images are stored for deploys.
    -y, --artif_payload_repo REPO    The name of the REPO where .tar.gz (service, meta, etcd2) files are stored.
	-p, --port PORT                  PORT to listen on (default: 8080).
    -L, --profiler_port PORT         *PORT the profiler is listening on (default: off).
    -X, --procs MAX                  *MAX processor cores to use from the machine.
    -D, --dsn DSN                    DSN string used to connect to database.

    -d, --debug                      Enable debugging output (default: false)

     *  Anything <= 0 is no change to the environment (default: 0).

Common options:
    -h, --help                       Show this message
    -V, --version                    Show version

Example:

    coreos-deploy -N "San Francisco" -H 0.0.0.0 -O example.com -E development \
	  -s http://dev-coreos.example.com:80 -k D3Pl0YT0Ken \
	  -a https://example.artifactoryonline.com/exampletest/api \
	  -u sysadm -w letmein -g 600 -t docker-v2-local-dev -y payload-v2-local-dev \
	  -p 8080 -X 2 --dsn "id:password@tcp(your-amazonaws-uri.com:3306)/dbname"

	for DSN usage, see https://github.com/go-sql-driver/mysql
`

// PrintUsageAndExit is used to print out command line options.
func PrintUsageAndExit() {
	fmt.Printf("%s\n", usageText)
	os.Exit(0)
}
