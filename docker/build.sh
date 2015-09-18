#!/bin/bash
docker build -t composer22/coreos-artifactory-monitor_build .
docker run -v /var/run/docker.sock:/var/run/docker.sock -v $(which docker):$(which docker) -ti --name coreos-artifactory-monitor_build composer22/coreos-artifactory-monitor_build
docker rm coreos-artifactory-monitor_build
docker rmi composer22/coreos-artifactory-monitor_build
