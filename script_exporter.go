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
	Output   string
}

type OutputType string

const (
	Number OutputType = "number"
)

func runScript(script *Script) (*bytes.Buffer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(script.Timeout)*time.Second)
	defer cancel()

	bashCmd := exec.CommandContext(ctx, *shell)

	bashIn, err := bashCmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	var bashOut bytes.Buffer;

	if script.Output != "" {
		bashCmd.Stdout = &bashOut
	}

	if err = bashCmd.Start(); err != nil {
		return nil, err
	}

	if _, err = bashIn.Write([]byte(script.Content)); err != nil {
		return &bashOut, err
	}

	bashIn.Close()

	return &bashOut, bashCmd.Wait()
}

func runScripts(scripts []*Script) []*Measurement {
	measurements := make([]*Measurement, 0)

	ch := make(chan *Measurement)

	for _, script := range scripts {
		go func(script *Script) {
			start := time.Now()
			success := 0
			status := -1
			output := ""
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

			if outBuffer != nil {
				output = outBuffer.String()
			}

			ch <- &Measurement{
				Script:   script,
				Duration: duration,
				Success:  success,
				Status:   status,
				Output:   output,
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

func processNumberOutput(w http.ResponseWriter, measurement *Measurement) {
	output, err := strconv.ParseFloat(measurement.Output, 64)
	if err != nil {
		log.Errorf("Error parsing number from script output: %s", err)
		return
	}
	fmt.Fprintf(w, "script_output{script=\"%s\"} %f\n", measurement.Script.Name, output)
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

		switch measurement.Script.Output {
		case string(Number):
			processNumberOutput(w, measurement)
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
