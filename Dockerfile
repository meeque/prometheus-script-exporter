FROM golang@sha256:f85330846cde1e57ca9ec309382da3b8e6ae3ab943d2739500e08c86393a21b1 AS build-env

RUN apk add --update git gcc libc-dev

RUN mkdir -p /go/script_exporter
COPY go.mod go.sum *.go /go/script_exporter/

WORKDIR /go/script_exporter
RUN go build



FROM alpine@sha256:a2d49ea686c2adfe3c992e47dc3b5e7fa6e6b5055609400dc2acaeb241c829f4

LABEL upstream="https://github.com/meeque/prometheus-script-exporter"

RUN apk add --no-cache bash skopeo jq

COPY --from=build-env /go/script_exporter/script_exporter /bin/script-exporter
COPY script-exporter.yml /etc/script-exporter/config.yml

EXPOSE 9172
ENTRYPOINT [ "/bin/script-exporter" ]
CMD [ "-config.file=/etc/script-exporter/config.yml" ]
