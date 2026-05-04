# Meeque's Fork of Prometheus Script Exporter

This is a fork of the [official](https://prometheus.io/docs/instrumenting/exporters/) [Prometheus Script Exporter](https://github.com/adhocteam/script_exporter).
It is intended for personal use, and may not be ready for general production use.
That said, let me know if you want to use it and have improvement suggestions.

Like the original, this Prometheus exporter is written to execute and collect metrics on script execution.
Unlike the original, it exposes the actual result status of the executed script, not just a boolean flag indicating success.

It also supports parsing the script outputs and export metrics based on this.
For now, it supports the following ways of exporting metrics based on script outputs:

1. Parse the complete output of the script as a `number` and export it as a single metric.
2. Parse the output as `json` and export each numeric json value as a distinct metris.

It only processes outputs of scripts that are configured for it, see config samples below.

Recommended Go Version: 1.24.4 or higher

## Sample Configuration

```yaml
scripts:
  - name: success
    script: sleep 5

  - name: failure
    script: sleep 2 && exit 1

  - name: timeout
    script: sleep 5
    timeout: 1

  - name: 'number'
    script: echo 23
    output: number

  - name: 'json'
    script: |
      echo '{"foo": 42, "bar": 2.71828}'
    output: json
```

## Running

You can run via docker with:

```
docker run -d -p 9172:9172 --name script-exporter \
  -v `pwd`/script-exporter.yml:/etc/script-exporter/config.yml:ro \
  meeque/script-exporter:latest \
  -config.file=/etc/script-exporter/config.yml \
  -web.listen-address=":9172" \
  -web.telemetry-path="/metrics" \
  -config.shell="/bin/sh"
```

You'll need to customize the docker image or use the binary on the host system
to install tools such as curl for certain scenarios.

## Probing

To return the script exporter internal metrics exposed by the default Prometheus
handler:

`$ curl http://localhost:9172/metrics`

To execute a script, use the `name` parameter to the `/probe` endpoint:

`$ curl http://localhost:9172/probe?name=failure`

```
script_duration_seconds{script="failure"} 2.006982
script_status{script="failure"} 1
script_success{script="failure"} 0
```

A regular expression may be specified with the `pattern` paremeter:

`$ curl http://localhost:9172/probe?pattern=.*`

```
script_duration_seconds{script="timeout"} 1.002526
script_status{script="timeout"} -1
script_success{script="timeout"} 0
script_duration_seconds{script="failure"} 2.004856
script_status{script="failure"} 1
script_success{script="failure"} 0
script_duration_seconds{script="success"} 5.006726
script_status{script="success"} 0
script_success{script="success"} 1
```

## Design

YMMV if you're attempting to execute a large number of scripts, and you'd be
better off creating an exporter that can handle your protocol without launching
shell processes for each scrape.
