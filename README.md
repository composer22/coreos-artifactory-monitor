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
    -t, --art_image_repo REPO        The name of the REPO where Docker images are stored for deploys.
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
	  -u sysadm -w letmein -g 600 -t docker-v2-local-dev -y payload-v2-local-dev \
	  -p 8080 -X 2 --dsn "id:password@tcp(your-amazonaws-uri.com:3306)/dbname"

	for DSN usage, see https://github.com/go-sql-driver/mysql
```
Please also see /docker dir for more information on running this service.

## Artifactory setup instructions

This section give information on the naming conventions and structure of the repo and
assets associated with the deploy process, and what the server is expecting while monitoring
artifactory.

### Repository structure

Two repositories should be setup in artifactory. One should contain the Docker images
that need to be deployed for an environment. The other contains .tar files for each version and environment.
The service monitors the docker repository for any application changes, such as new docker images or version
modification. If a change is noticed, it will look in the corresponding payload repo for a .tar file
for the application and version. The tar file contains metadata, etcd2 keys, and the service unit file for the
CoresOS environment.

The repositories on artifactory should follow the following best practice pattern.

```
# Docker repository:

/docker-repo-name-env/
  /docker-appimage-name
     /version-tag-A
	 /version-tag-B
	 ...

# Payload repository:
/payload-repo-name/
  /appimage-name
     /domain-environment-appname-versiontagA.tar.gz
	 /domain-environment-appname-versiontagB.tar.gz
	 ...
```
...for example:
```
# Docker repository:

/dockerv2-local-development/
  /video-mobile
     /1.0.1-21
	 /1.0.1-22

# Payload repository:
/payloadv2-local/
  /video-mobile
     /example.com-development-video-mobile-1.0.1-21.tar.gz
	 /example.com-development-video-mobile-1.0.1-22.tar.gz
	 ...
```
### Payload requirements

The .tar should contain ONLY the following files:

* metadata file (required) - a json file describing the deploy need for that version in a particular environment.
* service file(required) - a fleetctl .service or .service.tmpl file that describes how to launch the docker image in the cluster.
* etcd2 file (optional) - a file containing etc2 key/values that need to be added or changed for this version to work in the environment.
* README.md (optional) - a file describing the release and any notes.

### Metadata file

The metadata file should contain the following json attributes and should have a .json extention. Only one .json
file should be in the tar.gz.

* name - the name of the application being deployed. This should match the name of the service unit file.
* version - the version of the service to deploy. This is used when launching the service unit/template.
* imageVersion - should match the Docker image.
* numInstance - the number of coreos units to launch in the cluster.

Example:

```
video-mobile.json
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
/example.com-development-video-mobile-1.0.1-22/
  video-mobile-1.0.1-22@.service
  video-mobile-1.0.1-22@.service.tmpl
  video-mobile-1.0.1-22.etcd2
```
For more tech detail and examples, such as the template mechanism provided by coreos-client library,
please see the [coreos-deploy](http://github.com/composer22/coreos-deploy) and [coreos-deploy-client](http://github.com/composer22/coreos-deploy-client) projects.

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
