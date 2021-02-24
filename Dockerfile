FROM golang:alpine as builder

WORKDIR $GOPATH/src/mypackage/myapp/

# use modules
COPY go.mod .

ENV GO111MODULE=on
RUN go mod download
RUN go mod verify

COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -a -installsuffix cgo -o /go/bin/configmap-to-http .

FROM scratch

COPY --from=builder /go/bin/configmap-to-http /go/bin/configmap-to-http

ENTRYPOINT ["/go/bin/configmap-to-http"]