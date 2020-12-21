FROM golang:1.15 AS builder
WORKDIR /app
COPY . .
RUN make

###############################################################################

FROM ubuntu:focal
RUN apt update -y
RUN apt install iputils-ping iproute2 -y

WORKDIR /agent
COPY --from=builder /app/latency-sidecar .
CMD ./latency-sidecar
