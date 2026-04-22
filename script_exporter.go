package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
)

var (
	showVersion   = flag.Bool("version", false, "Print version information.")
	configFile    = flag.String("config.file", "script-exporter.yml", "Script exporter configuration file.")
	listenAddress = flag.String("web.listen-address", ":9172", "The address to listen on for HTTP requests.")
	metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	shell         = flag.String("config.shell", "/bin/sh", "Shell to execute script")
)

type Config struct {
	Scripts []*Script `yaml:"scripts"`
}

type Script struct {
	Name    string `yaml:"name"`
	Content string `yaml:"script"`
	Timeout int64  `yaml:"timeout"`
	Output  string `yaml:"output,omitempty"`
}

type Measurement struct {
	Script   *Script
	Success  int
	Status   int
	Duration float64
	Output   *any
}

type OutputType string

const (
	Number OutputType = "number"
)

func processNumberOutput(output *bytes.Buffer) (result float64, err error) {
	trimmedOutput := strings.TrimSpace(output.String())
	result, err = strconv.ParseFloat(trimmedOutput, 64)
	return
}

func processOutput(script *Script, output *bytes.Buffer) (result *any, err error) {
	if output == nil {
		return
	}
	var res any
	switch script.Output {
	case string(Number):
		res, err = processNumberOutput(output)
		return &res, err
	default:
		return nil, errors.New("unsupported output type")
	}
}

func runScript(script *Script) (stdout *bytes.Buffer, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(script.Timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, *shell)

	cmd.Stdin = strings.NewReader(script.Content)

	if script.Output != "" {
		stdout = &bytes.Buffer{}
		cmd.Stdout = stdout
	}

	if err = cmd.Start(); err != nil {
		return
	}
	err = cmd.Wait()
	return
}

func runScripts(scripts []*Script) []*Measurement {
	measurements := make([]*Measurement, 0)

	ch := make(chan *Measurement)

	for _, script := range scripts {
		go func(script *Script) {
			start := time.Now()
			success := 0
			status := -1
			outBuffer, err := runScript(script)
			duration := time.Since(start).Seconds()

			if err == nil {
				log.Debugf("OK: %s (after %fs).", script.Name, duration)
				success = 1
				status = 0
			} else {
				log.Infof("ERROR: %s: %s (failed after %fs).", script.Name, err, duration)
				var exitErr *exec.ExitError
				if errors.As(err, &exitErr) {
					status = exitErr.ExitCode()
				}
			}

			processedOutput, err := processOutput(script, outBuffer)
			if err != nil {
				log.Infof("ERROR: %s: failed processing script output as %s: %s", script.Name, script.Output, err)
			}

			ch <- &Measurement{
				Script:   script,
				Duration: duration,
				Success:  success,
				Status:   status,
				Output:   processedOutput,
			}
		}(script)
	}

	for i := 0; i < len(scripts); i++ {
		measurements = append(measurements, <-ch)
	}

	return measurements
}

func scriptFilter(scripts []*Script, name, pattern string) (filteredScripts []*Script, err error) {
	if name == "" && pattern == "" {
		err = errors.New("`name` or `pattern` required")
		return
	}

	var patternRegexp *regexp.Regexp

	if pattern != "" {
		patternRegexp, err = regexp.Compile(pattern)

		if err != nil {
			return
		}
	}

	for _, script := range scripts {
		if script.Name == name || (pattern != "" && patternRegexp.MatchString(script.Name)) {
			filteredScripts = append(filteredScripts, script)
		}
	}

	return
}

func scriptRunHandler(w http.ResponseWriter, r *http.Request, config *Config) {
	params := r.URL.Query()
	name := params.Get("name")
	pattern := params.Get("pattern")

	scripts, err := scriptFilter(config.Scripts, name, pattern)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	measurements := runScripts(scripts)

	for _, measurement := range measurements {
		fmt.Fprintf(w, "script_duration_seconds{script=\"%s\"} %f\n", measurement.Script.Name, measurement.Duration)
		fmt.Fprintf(w, "script_status{script=\"%s\"} %d\n", measurement.Script.Name, measurement.Status)
		fmt.Fprintf(w, "script_success{script=\"%s\"} %d\n", measurement.Script.Name, measurement.Success)

		if measurement.Output != nil {
			switch (*measurement.Output).(type) {
			case float64:
				fmt.Fprintf(w, "script_output{script=\"%s\"} %f\n", measurement.Script.Name, (*measurement.Output).(float64))
			}
		}
	}
}

func init() {
	prometheus.MustRegister(version.NewCollector("script_exporter"))
}

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Fprintln(os.Stdout, version.Print("script_exporter"))
		os.Exit(0)
	}

	log.Infoln("Starting script_exporter", version.Info())

	yamlFile, err := os.ReadFile(*configFile)

	if err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	config := Config{}

	err = yaml.Unmarshal(yamlFile, &config)

	if err != nil {
		log.Fatalf("Error parsing config file: %s", err)
	}

	log.Infof("Loaded %d script configurations", len(config.Scripts))

	for _, script := range config.Scripts {
		if script.Timeout == 0 {
			script.Timeout = 15
		}
	}

	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		scriptRunHandler(w, r, &config)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Script Exporter</title></head>
			<body>
			<h1>Script Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	log.Infoln("Listening on", *listenAddress)

	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		log.Fatalf("Error starting HTTP server: %s", err)
	}
}
