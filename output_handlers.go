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
	Process(output *bytes.Buffer) (result any, err error)
	Sample(scriptName string, result any) (samples []string)
}

type NumberOutputHandler struct {
}

func (NumberOutputHandler) Process(output *bytes.Buffer) (any, error) {
	trimmedOutput := strings.TrimSpace(output.String())
	return strconv.ParseFloat(trimmedOutput, 64)
}

func (NumberOutputHandler) Sample(scriptName string, result any) (samples []string) {
	sample := fmt.Sprintf("script_output{script=\"%s\"} %f", scriptName, result.(float64))
	return []string{sample}
}

type JsonOutputHandler struct {
}

func (JsonOutputHandler) Process(output *bytes.Buffer) (any, error) {
	var result any
	err := json.Unmarshal(output.Bytes(), &result)
	return result, err
}

func (JsonOutputHandler) Sample(scriptName string, result any) (samples []string) {
	outputs := map[string]string{}
	flattenJson(".", result, &outputs)
	for name, value := range outputs {
		samples = append(samples, fmt.Sprintf("script_output{script=\"%s\",output=\"%s\"} %s", scriptName, name, value))
	}
	return
}

func flattenJson(path string, value any, outputs *map[string]string) {
	switch value := value.(type) {
	case map[string]any:
		for k, v := range value {
			flattenJson(appendToPath(path, k), v, outputs)
		}
	case []any:
		for i, v := range value {
			flattenJson(appendToPath(path, fmt.Sprintf("%d", i)), v, outputs)
		}
	case float64:
		(*outputs)[path] = fmt.Sprintf("%f", value)
	case bool:
		numericValue := 0.0
		if value {
			numericValue = 1
		}
		flattenJson(path, numericValue, outputs)
	case string:
		trimmedValue := strings.TrimSpace(value)
		if numericValue, err := strconv.ParseFloat(trimmedValue, 64); err == nil {
			flattenJson(path, numericValue, outputs)
		} else {
			flattenJson(path, nil, outputs)
		}
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
