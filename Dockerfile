FROM golang:1.8.3-alpine3.6

WORKDIR /go/src/github.com/dhiltgen/docker-ee-chargeback

# TODO consider vendoring...
RUN apk add --update git && go get \
    github.com/Sirupsen/logrus\
    github.com/codegangsta/cli \
    github.com/prometheus/client_golang/api \
    github.com/prometheus/common/model

COPY . /go/src/github.com/dhiltgen/docker-ee-chargeback

RUN go build -o /go/bin/chargeback ./main/main.go


FROM alpine:3.6
COPY --from=0 /go/bin/chargeback /bin/chargeback

ENTRYPOINT ["/bin/chargeback"]
