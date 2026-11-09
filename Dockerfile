FROM golang@sha256:dbaa054703263d06c8b49576667418473b6231f8c0fa69dee103207fdb8e3654 AS build-env

RUN apk add --update git gcc libc-dev

RUN mkdir -p /go/script_exporter
COPY go.mod go.sum *.go /go/script_exporter/

WORKDIR /go/script_exporter
RUN go build



FROM alpine@sha256:3f4d6b32e26703c6082839234f43769d5046b3ac7c910033911d1e71d8cd630a

LABEL upstream="https://github.com/meeque/prometheus-script-exporter"

RUN apk add --no-cache bash skopeo jq

COPY --from=build-env /go/script_exporter/script_exporter /bin/script-exporter
COPY script-exporter.yml /etc/script-exporter/config.yml

EXPOSE 9172
ENTRYPOINT [ "/bin/script-exporter" ]
CMD [ "-config.file=/etc/script-exporter/config.yml" ]
