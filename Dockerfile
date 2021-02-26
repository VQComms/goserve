FROM golang:latest as builder

WORKDIR /go/src/vqcomms/goServe/

# use modules
COPY go.mod .

ENV GO111MODULE=on
RUN go mod download
RUN go mod verify

COPY . .

# Build the binary
RUN mkdir ./bin && \
    CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -tags netgo -installsuffix netgo -o ./bin/goServe && \
    mkdir ./bin/etc && \
    ID=$(shuf -i 100-9999 -n 1) && \
    echo $ID && \
    echo "appuser:x:$ID:$ID::/sbin/nologin:/bin/false" > ./bin/etc/passwd && \
    echo "appgroup:x:$ID:appuser" > ./bin/etc/group

FROM scratch

WORKDIR /

COPY --from=builder /go/src/vqcomms/goServe/bin .

ENTRYPOINT ["/goServe"]