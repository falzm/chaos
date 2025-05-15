FROM golang:latest AS builder
COPY . .
RUN GCO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /chaos-proxy ./cmd/chaos-proxy && go build -o /chaosctl ./cmd/chaosctl

FROM busybox:latest
COPY --from=builder /chaos-proxy /chaos-proxy
COPY --from=builder /chaosctl /chaosctl
EXPOSE 8000
EXPOSE 8666
ENTRYPOINT ["/chaos-proxy"]
