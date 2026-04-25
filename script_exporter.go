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
	Name    string     `yaml:"name"`
	Content string     `yaml:"script"`
	Timeout int64      `yaml:"timeout"`
	Output  OutputType `yaml:"output,omitempty"`
}

func executeScript(script *Script) (stdout *bytes.Buffer, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(script.Timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, *shell)

	cmd.Stdin = strings.NewReader(script.Content)

	if _, ok := outputHandlers[script.Output]; ok {
		stdout = &bytes.Buffer{}
		cmd.Stdout = stdout
	}

	if err = cmd.Start(); err != nil {
		return
	}
	err = cmd.Wait()
	return
}

func runScript(script *Script) (samples []string, err error) {

	start := time.Now()
	success := 0
	status := -1
	outBuffer, err := executeScript(script)
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

	samples = []string{}
	samples = append(samples, fmt.Sprintf("script_duration_seconds{script=\"%s\"} %f", script.Name, duration))
	samples = append(samples, fmt.Sprintf("script_status{script=\"%s\"} %d", script.Name, status))
	samples = append(samples, fmt.Sprintf("script_success{script=\"%s\"} %d", script.Name, success))

	handler, outputHandlerOk := outputHandlers[script.Output]
	if outputHandlerOk {
		handlerSamples := handler.Handle(script.Name, outBuffer)
		samples = append(samples, handlerSamples...)
	}
	return
}

func runScripts(scripts []*Script) (samples []string) {

	ch := make(chan []string, len(scripts))

	for _, script := range scripts {
		go func(script *Script) {
			// XXX check how errors should be processed! should runScripts return err at all?
			samples, _ := runScript(script)
			ch <- samples
		}(script)
	}

	for i := 0; i < len(scripts); i++ {
		samples = append(samples, <-ch...)
	}

	return samples
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

	samples := runScripts(scripts)
	for _, sample := range samples {
		fmt.Fprintln(w, sample)
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
