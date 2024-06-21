FROM 714967089364.dkr.ecr.ca-central-1.amazonaws.com/golang:1.22.4

WORKDIR /go/

ENV GOPATH /go/src
ENV GOBIN /go/bin

COPY ./cmd ./cmd/
COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum

# Download all dependencies and generate the binary
RUN go mod download && \
    go install ./cmd/cache-invalidator-api && \
    chmod -R +x ./bin

CMD ["./bin/cache-invalidator-api"]