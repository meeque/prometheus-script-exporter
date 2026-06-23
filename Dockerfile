FROM golang@sha256:3ad57304ad93bbec8548a0437ad9e06a455660655d9af011d58b993f6f615648 AS build-env

RUN apk add --update git gcc libc-dev

RUN mkdir -p /go/script_exporter
COPY go.mod go.sum *.go /go/script_exporter/

WORKDIR /go/script_exporter
RUN go build



FROM alpine@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b

LABEL upstream="https://github.com/meeque/prometheus-script-exporter"

RUN apk add --no-cache bash skopeo jq

COPY --from=build-env /go/script_exporter/script_exporter /bin/script-exporter
COPY script-exporter.yml /etc/script-exporter/config.yml

EXPOSE 9172
ENTRYPOINT [ "/bin/script-exporter" ]
CMD [ "-config.file=/etc/script-exporter/config.yml" ]
