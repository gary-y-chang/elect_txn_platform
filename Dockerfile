# build stage
FROM golang:1.11.5-alpine AS build-env
ADD . /go/src/gitlab.com/wondervoyage/platform
WORKDIR /go/src/gitlab.com/wondervoyage/platform
RUN apk add --update git
ENV GO111MODULE=on
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install

# final stage
FROM alpine
WORKDIR /goapp
ADD ./rest/api-doc/apidoc.yaml /goapp/apidoc.yaml
ADD ./configs/env-config.yaml /goapp/env-config.yaml
ADD ./configs/fabric/staging/config-docker.yaml /goapp/fabric/config.yaml
ADD ./configs/fabric/staging/crypto-config/ /goapp/fabric/crypto-config/
COPY --from=build-env /go/bin/platform /goapp
ENTRYPOINT ./platform