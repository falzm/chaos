# chaos-proxy

`chaos-proxy` is a basic *reverse-proxy* embedding a [Chaos middleware](https://github.com/falzm/chaos) instance. It can
be used in front of a back-end HTTP service to inject chaos when it's not possible to implement the Go HTTP middleware
natively.

## Installation

```
go get -u github.com/falzm/chaos/cmd/chaos-proxy
```

## Example Usage

```
chaos-proxy \
	-bind-addr 127.0.0.1:8001 \
	-controller-bind-addr unix:/var/run/chaos.sock \
	-url http://localhost:8000
```
