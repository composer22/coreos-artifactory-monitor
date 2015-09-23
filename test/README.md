## Additional Test Scripts and Applications

This folder contains scripts and files for testing requests against the server from the client side.

### Rested.App scripts - see: http://www.helloresolven.com/portfolio/rested/

./rested/*

- Health.request - Validate Health ping is returning 200 OK.
- Info.request - Receives information on the server and 200 OK.
- Metrics.request - Receives information and metrics on the server and 200 OK.
- Force.request - Used to force the server to check for deploy requests on Artifactory rather than wait on a timer event 200 OK.

