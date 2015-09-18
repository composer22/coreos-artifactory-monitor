### [Dockerized] (http://www.docker.com) [coreos-artifactory-monitor](https://registry.hub.docker.com/u/composer22/coreos-artifactory-monitor/)

A docker image for coreos-artifactory-monitor. This is created as a single "static" executable using a lightweight image.

To make:

cd docker
./build.sh

Once it completes, you can run the server. For example:

docker run -v /tmp:/tmp --name=coreos_artifactory_monitor \
 -p 0.0.0.0:8080:8080 -h `hostname` \
  composer22/coreos-artifactory-monitor \
 -p 8080 \
 -H 0.0.0.0 \
 -X 2  \
 -s http://dev-coreos.example.com:80 \
 -k D3Pl0YT0Ken \
 -a https://example.artifactoryonline.com/exampletest/api \
 -u sysadm \
 -w letmein \
 -g 600 \
 -t docker-v2-local-dev \
 -y payload-v2-local-dev \
 -N "San Francisco" \
 -O example.com \
 -E development \
 --dsn  "id:password@tcp(your-amazonaws-uri.com:3306)/dbname"

see composer22/coreos-artifactory-monitor/coreos-artifactory-monitor.service for an example of
how to write your own service file for your CoreOS cluster.
