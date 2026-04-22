package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type OutputType string

const (
	Number OutputType = "number"
	Json   OutputType = "json"
)

type OutputHandler interface {
	Process(output *bytes.Buffer) (result any, err error)
	Print(writer io.Writer, scriptName string, result any)
}

type NumberOutputHandler struct {
}

func (h *NumberOutputHandler) Process(output *bytes.Buffer) (any, error) {
	trimmedOutput := strings.TrimSpace(output.String())
	return strconv.ParseFloat(trimmedOutput, 64)
}

func (h *NumberOutputHandler) Print(writer io.Writer, scriptName string, result any) {
	fmt.Fprintf(writer, "script_output{script=\"%s\"} %f\n", scriptName, result.(float64))
}

type JsonOutputHandler struct {
}

func (h *JsonOutputHandler) Process(output *bytes.Buffer) (any, error) {
	var result any
	err := json.Unmarshal(output.Bytes(), &result)
	return result, err
}

func (h *JsonOutputHandler) Print(writer io.Writer, scriptName string, result any) {
	for name, value := range result.(map[string]any) {
		fmt.Fprintf(writer, "script_output{script=\"%s\",output=\"%s\"} %f\n", scriptName, name, value)
	}
}

var outputHandlers = map[OutputType]OutputHandler{
	Number: &NumberOutputHandler{},
	Json:   &JsonOutputHandler{},
}
