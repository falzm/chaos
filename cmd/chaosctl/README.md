# chaosctl

`chaosctl` is a convenience utility to dynamically interact with a [Chaos middleware](https://github.com/falzm/chaos)
instance.

## Installation

```
go get -u github.com/falzm/chaos/cmd/chaosctl
```

## Example Usage

```
chaosctl add POST /api/a \
	--delay-duration 3000 \
	--delay-probability 0.5

chaosctl del POST /api/a
```

See `chaosctl --help` for detailed CLI usage.
