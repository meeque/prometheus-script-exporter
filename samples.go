package main

import (
	"bytes"
	"maps"
	"math"
	"slices"
	"strconv"
	"strings"
)

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

func (s1 Sample) Equal(s2 Sample) bool {
	return s1.EqualNameAndLabels(s1) &&
		(s1.Value == s2.Value || (math.IsNaN(s1.Value) && math.IsNaN(s2.Value)))
}

func (s1 Sample) EqualNameAndLabels(s2 Sample) bool {
	return s1.Name == s2.Name && maps.Equal(s1.Labels, s2.Labels)
}

func (s *Sample) String() string {
	return s.StringNameAndLabels() +
		" " +
		strconv.FormatFloat(s.Value, 'f', -1, 64)
}

func (s *Sample) StringNameAndLabels() string {
	buf := bytes.NewBufferString(s.Name)
	buf.WriteRune('{')
	labelNames := slices.Collect(maps.Keys(s.Labels))
	slices.Sort(labelNames)
	for labelNum, labelName := range labelNames {
		buf.WriteString(encodeSamplePart(labelName, false))
		buf.WriteRune('=')
		buf.WriteString(encodeSamplePart(s.Labels[labelName], true))
		if labelNum < len(labelNames)-1 {
			buf.WriteRune(',')
		}
	}
	buf.WriteRune('}')
	return buf.String()
}

func encodeSamplePart(part string, force bool) string {
	if force || strings.ContainsAny(part, "{},=\n\"\\") {
		quotedPart := ""
		for _, r := range part {
			switch r {
			case '\n':
				quotedPart += `\n`
			case '"':
				quotedPart += `\"`
			case '\\':
				quotedPart += `\\`
			default:
				quotedPart += string(r)
			}
		}
		return `"` + quotedPart + `"`
	}
	return part
}
