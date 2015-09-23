# coreos-artifactory-monitor
[![License MIT](https://img.shields.io/npm/l/express.svg)](http://opensource.org/licenses/MIT)
[![Build Status](https://travis-ci.org/composer22/coreos-artifactory-monitor.svg?branch=master)](http://travis-ci.org/composer22/coreos-artifactory-monitor)
[![Current Release](https://img.shields.io/badge/release-v0.0.1-brightgreen.svg)](https://github.com/composer22/coreos-artifactory-monitor/releases/tag/v0.0.1)
[![Coverage Status](https://coveralls.io/repos/composer22/coreos-artifactory-monitor/badge.svg?branch=master)](https://coveralls.io/r/composer22/coreos-artifactory-monitor?branch=master)

A service to monitor artifactory for Docker deploys written in [Go.](http://golang.org)

## About

This service monitors artifactory repository folder for docker image version changes. If one is
detected, the service will deploy the new or newer version to a given CoreOS cluster by submitting
a request to the coreos-deploy service running within that cluster.

## Requirements

A MySQL database is required. For the DB schema, please see ./db/schema.sql

## CLI Usage

```
Description: coreos-artifactory-monitor is a server for monitoring deploy needs from Artifactory to a coreos cluster.

Usage: coreos-artifactory-monitor [options...]

Server options:
    -N, --name NAME                  NAME of the server (default: empty field).
    -H, --hostname HOSTNAME          HOSTNAME of the server (default: localhost).
    -O, --domain DOMAIN              DOMAIN of the site being managed (default: localhost).
    -E, --environment ENVIRONMENT    ENVIRONMENT (development, qa, staging, production).
    -s, --deploy_url URL             URL to the coreos-deploy service.
    -k, --deploy_token TOKEN         Security TOKEN to access the coreos-deploy service.
    -a, --art_endpoint APIURL        The base APIURL to the artifactory API service.
    -u, --art_userid USERID          USERID to login to the artifactory API service.
    -w, --art_password PASSWORD      PASSWORD to login to the artifactory API service.
    -g, --art_polling INTERVAL       How often to check artifactory for deploys in INTERVAL seconds (default: 300 sec).
    -t, --art_deploy_repo REPO       The name of the REPO where the deploy request files are stored.
    -y, --art_payload_repo REPO      The name of the REPO where .tar.gz (service, meta, etcd2) files are stored.
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
	  -u sysadm -w letmein -g 600 -t cluster-deploys-dev -y cluster-payloads \
	  -p 8080 -X 2 --dsn "id:password@tcp(your-amazonaws-uri.com:3306)/dbname"

	for DSN usage, see https://github.com/go-sql-driver/mysql
```
Please also see /docker dir for more information on running this service.

## Artifactory setup instructions

This section give information on the naming conventions and structure of the repo and
assets associated with the deploy process, and what the server is expecting while monitoring
artifactory.

### Repository structure
Two repositories should be set up in Artifactor. The first should contains .tar files for each version and environment.
These tar files should contain the etcd2 keys and .service file and metadata. The second repo acts as a deploy
request front. The service monitors the second repository for any version changes within an application folder.
If a change is noticed, it will look in the corresponding payload repo for a .tar file for the application and version,
uncompress it, and use the information contained to submit the deploy request.

Optionally, you might consider using Artifactory as a store for Docker images. The .service file can be configured
to pull down Docker images from any repository, but we are assuming for this application you are using Artifactory.

The repositories on artifactory should follow the following best practice pattern.

```
# Deploy request repository:
/deploy-req-repo-name/
  /appimage-name
     /version-tag-A.deploy
	 /version-tag-B.deploy
     ...

# Payload repository:
/payload-repo-name/
  /appimage-name
     /domain-environment-appname-versiontagA.tar.gz
	 /domain-environment-appname-versiontagB.tar.gz
	 ...

# Docker repository:
/docker-repo-name/
  /appimage-name
     /version-tag-A
	 /version-tag-B
	 ...
```
...for example:
```
# Deploy request repository:
/cluster-deploys-development/
  /video-mobile
    /1.0.1-21.deploy
	/1.0.1-22.deploy
	 ...

# Payload repository:
/cluster-payloads/
  /video-mobile
     /example.com-development-video-mobile-1.0.1-21.tar.gz
	 /example.com-development-video-mobile-1.0.1-22.tar.gz
	 ...

# Docker repository:
/dockerv2-local/
  /video-mobile
     /1.0.1-21
	 /1.0.1-22
     ...
```
The .deploy file above is simply a "touched" file that indicates a version has been posted and ready to be deployed
to a particular environment. It is essentially empty. It might contain information in the future. The tar.gz files
contain all information related to the deploy for an environment.

### Payload requirements

The .tar should contain ONLY the following files:

* metadata file (1 required) - a json file describing the deploy need for that version in a particular environment.
* service file(1 required) - a fleetctl .service or .service.tmpl file that describes how to launch the docker image in the cluster.
* etcd2 file (1 optional) - a file containing etc2 key/values that need to be added or changed for this version to work in the environment.
* README.md (1 optional) - a file describing the release and any notes.

### Metadata file

The metadata file should contain the following json attributes and should have a .json extention. Only one .json
file should be in the tar.gz.

* name - the name of the application being deployed. This should match the name of the .service unit file.
* version - the version of the service to deploy. This is used when launching the service unit/template.
* imageVersion - should match the Docker image. Usually this is the same as 'version'
* numInstance - the number of coreos units to launch in the cluster for this environment.

Example:

```
video-mobile.metadata.json
{
  "name":"video-mobile",
  "version":"1.0.1-22",
  "imageVersion":"1.0.1-22",
  "numInstance": 2
}
```
### .service, .service.tmpl, and .etcd2 key files.

These files should follow these naming conventions although they can be named anything,
as long as there is only one file type within the tar.gz. They should untar into a directory
with the same name as the tar.gz. For example:
```
example.com-development-video-mobile-1.0.1-22.tar.gze
/example.com-development-video-mobile-1.0.1-22/
  example.com-development-video-mobile-1.0.1-22.etcd2
  example.com-development-video-mobile.metadata.json
  video-mobile-1.0.1-22@.service
  - or -
  video-mobile-1.0.1-22@.service.tmpl

  README.md
```
For more tech detail and examples, such as the template mechanism provided by coreos-client library,
please see the [coreos-deploy](http://github.com/composer22/coreos-deploy) and [coreos-deploy-client](http://github.com/composer22/coreos-deploy-client) projects.

## HTTP API

Header for services other than /health should contain:

* Accept: application/json
* Authorization: Bearer with token
* Content-Type: application/json

The bearer tokens are stored in artifactory_auth_tokens in the database.

Example cURL:

```
$ curl -i -H "Accept: application/json" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer S0M3B3EARERTOK3N" \
-X GET "http://0.0.0.0:8080/v1.0/info"

HTTP/1.1 200 OK
Content-Type: application/json;charset=utf-8
Date: Fri, 03 Apr 2015 17:29:17 +0000
Server: San Francisco
X-Request-Id: DC8D9C2E-8161-4FC0-937F-4CA7037970D5
Content-Length: 0
```

Three API routes are provided for service measurement:

* http://localhost:8080/v1.0/health - GET: Is the server alive?
* http://localhost:8080/v1.0/info - GET: What are the params of the server?
* http://localhost:8080/v1.0/metrics - GET: What performance and statistics are from the server?

Calling the following API will force the server to check for new deploys immediately
instead of waiting a polling interval set by -g or --art_polling:

* http://localhost:8080/v1.0/force - GET: Check for new deploys immediately. Don't wait.

## Building

This code currently requires version 1.42 or higher of Go.

You will need to install the following dependencies:
```
go get github.com/composer22/coreos-deploy
go get github.com/composer22/coreos-deploy-client
```
Information on Golang installation, including pre-built binaries, is available at
<http://golang.org/doc/install>.

Run `go version` to see the version of Go which you have installed.

Run `go build` inside the directory to build.

Run `go test ./...` to run the unit regression tests.

A successful build run produces no messages and creates an executable called `coreos-artifactory-monitor` in this
directory.

Run `go help` for more guidance, and visit <http://golang.org/> for tutorials, presentations, references and more.

## Docker Images

A prebuilt docker image is available at (http://www.docker.com) [coreos-artifactory-monitor](https://registry.hub.docker.com/u/composer22/coreos-artifactory-monitor/)

If you have docker installed, run:
```
docker pull composer22/coreos-artifactory-monitor:latest

or

docker pull composer22/coreos-artifactory-monitor:<version>

if available.
```
See /docker directory README for more information on how to run it.

You should run this docker container on the control or service part of your cluster.
NOTE: Only one instance should be run per cluster at any time.

## License

(The MIT License)

Copyright (c) 2015 Pyxxel Inc.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to
deal in the Software without restriction, including without limitation the
rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
sell copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
IN THE SOFTWARE.
