package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
)

const (
	Number OutputType = "number"
	Json   OutputType = "json"
)

type OutputType string

type ProcessOutput func(metricName string, output *bytes.Buffer) (samples *[]Sample, err error)

func ProcessNumberOutput(metricName string, output *bytes.Buffer) (samples *[]Sample, err error) {
	trimmedOutput := strings.TrimSpace(output.String())
	numberOutput, err := strconv.ParseFloat(trimmedOutput, 64)

	if err != nil {
		err = fmt.Errorf("%s: failed processing script output as number: %w", metricName, err)
		return
	}

	samples = &[]Sample{*NewNumberOutputSample(metricName, numberOutput)}
	return
}

func NewNumberOutputSample(script string, value float64) (sample *Sample) {
	sample = NewScriptSample("script_output", script, value)
	return
}

func ProcessJsonOutput(metricName string, output *bytes.Buffer) (samples *[]Sample, err error) {
	var jsonOutput any
	err = json.Unmarshal(output.Bytes(), &jsonOutput)
	if err != nil {
		err = fmt.Errorf("%s: failed processing script output as a JSON: %w", metricName, err)
		return
	}

	flatJsonOutput := &FlatJsonOutput{}
	flatJsonOutput.append(".", jsonOutput)

	samples = &[]Sample{}
	for outputName, value := range *flatJsonOutput {
		*samples = append(*samples, *NewJsonOutputSample(metricName, outputName, value))
	}
	return
}

func NewJsonOutputSample(script string, output string, value float64) (sample *Sample) {
	sample = NewScriptSample("script_output", script, value)
	sample.Labels["output"] = output
	return
}

type FlatJsonOutput map[string]float64

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
		(*o)[path] = value
	default:
		log.Printf("WARN: Silently ignoring non-numeric JSON value at path '%s'", path)
	}
}

func appendToPath(basePath string, segment string) (path string) {
	if basePath == "." {
		return segment
	}
	return fmt.Sprintf("%s.%s", basePath, segment)
}

var processOutputByType = map[OutputType]ProcessOutput{
	Number: ProcessNumberOutput,
	Json:   ProcessJsonOutput,
}
