FROM quay.io/lib/golang:1.25-bookworm as builder

WORKDIR /app
RUN apt-get update && apt-get install -y libvirt-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o kubevirt-libvirt-metrics-exporter

FROM quay.io/lib/debian:12-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends libvirt0 ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /
COPY --from=builder /app/kubevirt-libvirt-metrics-exporter .

ENTRYPOINT ["/kubevirt-libvirt-metrics-exporter"]
