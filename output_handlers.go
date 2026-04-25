package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/prometheus/common/log"
)

type OutputType string

const (
	Number OutputType = "number"
	Json   OutputType = "json"
)

type OutputHandler interface {
	Handle(metricName string, output *bytes.Buffer) (samples []string)
}

type NumberOutputHandler struct {
}

func (h NumberOutputHandler) Handle(metricName string, output *bytes.Buffer) (samples []string) {
	trimmedOutput := strings.TrimSpace(output.String())
	numberOutput, err := strconv.ParseFloat(trimmedOutput, 64)

	if err != nil {
		log.Infof("ERROR: %s: failed processing script output as number: %s", metricName, err)
		return
	}

	samples = append(samples, fmt.Sprintf("script_output{script=\"%s\"} %f", metricName, numberOutput))
	return
}

type JsonOutputHandler struct {
}

func (h JsonOutputHandler) Handle(metricName string, output *bytes.Buffer) (samples []string) {
	var jsonOutput any
	err := json.Unmarshal(output.Bytes(), &jsonOutput)

	if err != nil {
		log.Infof("ERROR: %s: failed processing script output as a JSON: %s", metricName, err)
	}

	flatJsonOutput := &FlatJsonOutput{}
	flatJsonOutput.append(".", jsonOutput)

	for name, value := range *flatJsonOutput {
		samples = append(samples, fmt.Sprintf("script_output{script=\"%s\",output=\"%s\"} %s", metricName, name, value))
	}
	return
}

type FlatJsonOutput map[string]string

func (o *FlatJsonOutput) append(path string, value any) {
	switch value := value.(type) {
	// most cases involve recursion!
	case map[string]any:
		for k, v := range value {
			o.append(appendToPath(path, k), v)
		}
	case []any:
		for i, v := range value {
			o.append(appendToPath(path, fmt.Sprintf("%d", i)), v)
		}
	case bool:
		numericValue := 0.0
		if value {
			numericValue = 1
		}
		o.append(path, numericValue)
	case string:
		trimmedValue := strings.TrimSpace(value)
		if numericValue, err := strconv.ParseFloat(trimmedValue, 64); err == nil {
			o.append(path, numericValue)
		} else {
			o.append(path, nil)
		}
	// recursion only terminates below here
	case float64:
		(*o)[path] = fmt.Sprintf("%f", value)
	default:
		log.Debugf("WARN: Silently ignoring non-numeric JSON value at path '%s'", path)
	}
}

func appendToPath(basePath string, segment string) (path string) {
	if basePath == "." {
		return segment
	}
	return fmt.Sprintf("%s.%s", basePath, segment)
}

var outputHandlers = map[OutputType]OutputHandler{
	Number: &NumberOutputHandler{},
	Json:   &JsonOutputHandler{},
}
