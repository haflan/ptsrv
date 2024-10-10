FROM docker.io/library/golang:alpine AS builder

COPY . /go/src
WORKDIR /go/src
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
RUN go build -o /go/bin/ptsrv

FROM docker.io/library/alpine:latest
COPY --from=builder /go/bin/ptsrv /usr/bin/

CMD ["/usr/bin/ptsrv"]
