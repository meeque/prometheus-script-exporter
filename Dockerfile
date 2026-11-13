FROM golang@sha256:eb61c3873c97416fff0116553fff65d67a3549f72e11d9758849a689cdd8b523 AS build-env

RUN apk add --update git gcc libc-dev

RUN mkdir -p /go/script_exporter
COPY go.mod go.sum *.go /go/script_exporter/

WORKDIR /go/script_exporter
RUN go build



FROM alpine@sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf

LABEL upstream="https://github.com/meeque/prometheus-script-exporter"

RUN apk add --no-cache bash skopeo jq

COPY --from=build-env /go/script_exporter/script_exporter /bin/script-exporter
COPY script-exporter.yml /etc/script-exporter/config.yml

EXPOSE 9172
ENTRYPOINT [ "/bin/script-exporter" ]
CMD [ "-config.file=/etc/script-exporter/config.yml" ]
