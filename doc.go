/*
Chaos is a HTTP Negroni middleware that can be used to inject chaotic behavior into your web application (such as
delays and errors) in a controlled and programmatic way. It can be useful in chaos engineering for testing a
distributed system resiliency, or to ensure application observability instrumentation is working as intended.

The Chaos Middleware is configurable on-the-fly via a dedicated management HTTP controller. For earch target route
(i.e. the actual HTTP endpoint that will be impacted by this middleware), it is possible to set a chaos specification
defining either or both a delay artificially stalling the request processing and an error terminating the request
processing with an arbitrary status code and optional message.

Configuration Routes

For every configuration route, the following URL parameters are mandatory:

	<method>: the HTTP method corresponding to the target route (e.g. "GET", "POST"...)
	<path>: the URL path corresponding to the target route, starting (e.g. "/api/a")

The available routes are:

	PUT /

Set the chaos specification for the corresponding target route. The request body format is JSON-formatted and its
content type must be "application/json":

	{
	  "error": {
	    "status_code": <int: HTTP status code to return for request termination>,
	    "message": "<string: optional message to return for request termination>",
	    "p": <float: probability between 0 and 1>
	  },
	  "delay": {
	    "duration": <int: delay duration in milliseconds>,
	    "p": <float: probability between 0 and 1>
	  }
	}

Upon successful request, a "204 No Content" status code is returned.

	GET /

Get the chaos specification currently set for the corresponding target route:

	DELETE /

Delete the chaos specification set for the corresponding target route.

Example Usage

Set a 3 seconds delay with a 50% probability and a 504 error with a 100% probability for target route "POST /api/a":

	curl -X PUT -H 'Content-Type: application/json' \
		-d '{"error":{"status_code":504,"p":1},"delay":{"duration":3000,"p":0.5}}' \
		'localhost:8666/?method=POST&path=/api/a'

Set a 599 error with message "oh noes" with a 10% probability for target route "GET /api/b":

	curl -X PUT -H 'Content-Type: application/json' \
		-d '{"error":{"status_code":599,"message":"oh noes","p":0.1}}' \
		'localhost:8666/?method=GET&path=/api/b'

Get the currently set chaos specification for the target route "GET /api/b":

	curl -i 'localhost:8666/?method=GET&path=/api/b'
	(returns Error: 599 "oh noes" (probability: 0.1))

Delete the currently set chaos specification for the target route "GET /api/b":

	curl -i -X DELETE 'localhost:8666/?method=GET&path=/api/b'

Note: requests affected by a chaos specification feature a X-Chaos-Injected-* HTTP header
describing the nature of the disruption. Example:

	X-Chaos-Injected-Delay: 3s (probability: 0.5)
	X-Chaos-Injected-Error: 504 (probability: 1.0)
*/
package chaos
