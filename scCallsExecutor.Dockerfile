FROM golang:1.20.7-bookworm AS builder
LABEL description="This Docker image builds the SC calls executor binary."

WORKDIR /kleverchain
COPY . .

RUN go mod tidy

WORKDIR /kleverchain/cmd/scCallsExecutor

RUN APPVERSION=$(git describe --tags --long --always | tail -c 11) && echo "package main\n\nfunc init() {\n\tappVersion = \"${APPVERSION}\"\n}" > local.go
RUN go mod tidy
RUN go build

FROM ubuntu:22.04 AS runner
LABEL description="This Docker image runs SC calls executor binary."

RUN apt-get update \
    && apt-get -y install git \
    && apt-get clean

COPY --from=builder /kleverchain/cmd/scCallsExecutor /kleverchain

WORKDIR /kleverchain

ENTRYPOINT ["./scCallsExecutor"]
