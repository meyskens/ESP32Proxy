FROM --platform=$BUILDPLATFORM golang:1.17-alpine as build

ARG TARGETPLATFORM
ARG BUILDPLATFORM

RUN apk add --no-cache git

COPY ./ /go/src/github.com/meyskens/esp32proxy

WORKDIR /go/src/github.com/meyskens/esp32proxy

RUN export GOARM=6 && \
    export GOARCH=amd64 && \
    if [ "$TARGETPLATFORM" == "linux/arm64" ]; then export GOARCH=arm64; fi && \
    if [ "$TARGETPLATFORM" == "linux/arm" ]; then export GOARCH=arm; fi && \
    go build -ldflags "-X main.version=$(git rev-parse --short HEAD)" ./cmd/esp32proxy/

FROM alpine:3.13

RUN apk add --no-cache ca-certificates

COPY --from=build /go/src/github.com/meyskens/esp32proxy/esp32proxy /usr/local/bin/

RUN mkdir /opt/esp32proxy
WORKDIR /opt/esp32proxy

COPY config.json /opt/esp32proxy/config.json

ENTRYPOINT [ "/usr/local/bin/esp32proxy" ]
CMD [ "host" ]