package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"maps"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"slices"
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
	Name    string     `yaml:"name"`
	Content string     `yaml:"script"`
	Timeout int64      `yaml:"timeout"`
	Output  OutputType `yaml:"output,omitempty"`
}

type Sample struct {
	Name   string
	Labels map[string]string
	Value  float64
}

func NewSample(name string, labels map[string]string, value float64) (sample *Sample) {
	return &Sample{
		Name:   name,
		Labels: labels,
		Value:  value,
	}
}

func NewScriptSample(name string, script string, value float64) (sample *Sample) {
	labels := map[string]string{"script": script}
	return NewSample(name, labels, value)
}

func (s *Sample) String() string {
	buf := bytes.NewBufferString(s.Name)
	buf.WriteString("{")
	labelNames := slices.Collect(maps.Keys(s.Labels))
	slices.Sort(labelNames)
	for labelNum, labelName := range labelNames {
		buf.WriteString(encodeSamplePart(labelName, false))
		buf.WriteString("=")
		buf.WriteString(encodeSamplePart(s.Labels[labelName], true))
		if labelNum < len(labelNames)-1 {
			buf.WriteString(",")
		}
	}
	buf.WriteString("} ")
	buf.WriteString(strconv.FormatFloat(s.Value, 'f', -1, 64))
	return buf.String()
}

func encodeSamplePart(part string, force bool) string {
	if force || strings.ContainsAny(part, "\n\"\\") {
		part = strings.ReplaceAll(part, "\n", "\\n")
		part = strings.ReplaceAll(part, "\"", "\\\"")
		part = strings.ReplaceAll(part, "\\", "\\\\")
		return "\"" + part + "\""
	}
	return part
}

func executeScript(script *string, timeout int64, captureOutput bool) (stdout *bytes.Buffer, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, *shell)

	cmd.Stdin = strings.NewReader(*script)

	if captureOutput {
		stdout = &bytes.Buffer{}
		cmd.Stdout = stdout
	}

	if err = cmd.Start(); err != nil {
		return
	}
	err = cmd.Wait()
	return
}

func runScript(script *Script) (samples *[]Sample) {
	success := 0
	status := -1
	processOutput := processOutputByType[script.Output]

	start := time.Now()
	outBuffer, err := executeScript(&script.Content, script.Timeout, processOutput != nil)
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

	samples = &[]Sample{
		*NewScriptSample("script_duration_seconds", script.Name, duration),
		*NewScriptSample("script_status", script.Name, float64(status)),
		*NewScriptSample("script_success", script.Name, float64(success)),
	}

	if processOutput != nil {
		outputSamples, err := processOutput(script.Name, outBuffer)
		if err != nil {
			log.Infof("Silently ignoring error in %T: %s", processOutput, err)
			return
		}
		*samples = append(*samples, *outputSamples...)
	}

	return
}

func runScripts(scripts []*Script) (samples *[]Sample) {

	ch := make(chan []Sample, len(scripts))

	for _, script := range scripts {
		go func() {
			ch <- *runScript(script)
		}()
	}

	samples = &[]Sample{}
	for range scripts {
		*samples = append(*samples, <-ch...)
	}
	return
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
	for _, sample := range *samples {
		fmt.Fprintln(w, sample.String())
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
