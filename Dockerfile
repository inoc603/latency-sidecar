FROM ubuntu:focal

RUN apt update -y

RUN apt install iputils-ping iproute2 -y

ADD latency-sidecar .

CMD ./latency-sidecar
