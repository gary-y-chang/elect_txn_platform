# build stage
FROM golang AS build-env
ADD . /go/src/gitlab.com/wondervoyage/platform
WORKDIR /go/src/gitlab.com/wondervoyage/platform
RUN go get -u github.com/labstack/echo/...
RUN go get -u github.com/jinzhu/gorm/dialects/postgres
RUN go get -u github.com/jinzhu/gorm
RUN go get -u github.com/satori/go.uuid
RUN go get -u github.com/hyperledger/fabric-sdk-go/pkg/core/config
RUN go get -u github.com/hyperledger/fabric-sdk-go/pkg/fabsdk
RUN go get -u github.com/hyperledger/fabric-sdk-go/pkg/client/channel
RUN go get -u github.com/hyperledger/fabric-sdk-go/pkg/client/msp
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install

# final stage
FROM alpine
WORKDIR /goapp
COPY --from=build-env /go/bin/platform /goapp
ENTRYPOINT ./platform