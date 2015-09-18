package server

import "time"

const (
	version                = "0.0.1"       // Application and server version.
	DefaultHostName        = "localhost"   // The hostname of the server.
	DefaultEnvironment     = "development" // The default environment for the server.
	DefaultPort            = 8080          // Port to receive requests: see IANA Port Numbers.
	DefaultProfPort        = 0             // Profiler port to receive requests.*
	DefaultMaxProcs        = 0             // Maximum number of computer processors to utilize.*
	DefaultPollingInterval = 300           // Polling interval in seconds to check artifactory (5 min).

	// * zeros = no change or no limitations or not enabled.

	// http: routes.
	httpRouteV1Health  = "/v1.0/health"
	httpRouteV1Info    = "/v1.0/info"
	httpRouteV1Metrics = "/v1.0/metrics"
	httpRouteV1Force   = "/v1.0/force"

	// Artifactory API routes
	artSourceRoute = "/storage"

	// Connections.
	TCPReadTimeout  = 10 * time.Second
	TCPWriteTimeout = 10 * time.Second

	httpGet    = "GET"
	httpPost   = "POST"
	httpPut    = "PUT"
	httpDelete = "DELETE"
	httpHead   = "HEAD"
	httpTrace  = "TRACE"
	httpPatch  = "PATCH"

	// Directories and add ons for deploy work.
	tmpDir             = "/tmp/coreos-artifactory-monitor/"
	maxPollStatusCount = 6  // 6 times
	maxPollStatusPause = 10 // 10 seconds

	// Error messages.
	InvalidMediaType     = "Invalid Content-Type or Accept header value."
	InvalidMethod        = "Invalid Method for this route."
	InvalidBody          = "Invalid body of text in request."
	InvalidJSONText      = "Invalid JSON format in text of body in request."
	InvalidJSONAttribute = "Invalid - 'text' attribute in JSON not found."
	InvalidAuthorization = "Invalid authorization."
)
