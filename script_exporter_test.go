package main

import (
	"errors"
	"maps"
	"math"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func (s1 Sample) Equal(s2 Sample) bool {
	return s1.Name == s2.Name &&
		maps.Equal(s1.Labels, s2.Labels) &&
		(s1.Value == s2.Value || (math.IsNaN(s1.Value) && math.IsNaN(s2.Value)))
}

type SampleAsserter interface {
	Assert(t *testing.T, sample *Sample) (match bool)
}

type ExactMatchAsserter struct {
	Sample *Sample
}

func (a ExactMatchAsserter) Assert(t *testing.T, sample *Sample) (match bool) {
	return a.Sample.Equal(*sample)
}

type MinDurationAsserter struct {
	Name   string
	Labels map[string]string
	Min    float64
	Max    float64
}

func (a MinDurationAsserter) Assert(t *testing.T, sample *Sample) (match bool) {
	match = a.Name == sample.Name &&
		reflect.DeepEqual(a.Labels, sample.Labels)
	if match {
		if sample.Value < a.Min || sample.Value > a.Max {
			t.Errorf("Expected sampled duration to be between %f and %f, but got %f", a.Min, a.Max, sample.Value)
		}
	}
	return
}

var config = &Config{
	Scripts: []*Script{
		{"success", "exit 0", 15, ""},
		{"failure", "exit 1", 15, ""},
		{"timeout", "sleep 3", 1, ""},
		{"number", "echo 23", 15, "number"},
		{"json", "echo '{\"foo\": 42, \"bar\": 2.71828}'", 15, "json"},
	},
}

func TestRunScripts(t *testing.T) {
	expectedSamples := []any{
		MinDurationAsserter{
			"script_duration_seconds",
			map[string]string{"script": "success"},
			0.0,
			0.5,
		},
		"script_status{script=\"success\"} 0",
		"script_success{script=\"success\"} 1",

		MinDurationAsserter{
			"script_duration_seconds",
			map[string]string{"script": "failure"},
			0.0,
			0.5,
		},
		"script_status{script=\"failure\"} 1",
		"script_success{script=\"failure\"} 0",

		MinDurationAsserter{
			"script_duration_seconds",
			map[string]string{"script": "timeout"},
			0.9,
			1.4,
		},
		"script_status{script=\"timeout\"} -1",
		"script_success{script=\"timeout\"} 0",

		MinDurationAsserter{
			"script_duration_seconds",
			map[string]string{"script": "number"},
			0.0,
			0.5,
		},
		"script_status{script=\"number\"} 0",
		"script_success{script=\"number\"} 1",
		"script_output{script=\"number\"} 23",

		MinDurationAsserter{
			"script_duration_seconds",
			map[string]string{"script": "json"},
			0.0,
			0.5,
		},
		"script_status{script=\"json\"} 0",
		"script_success{script=\"json\"} 1",
		"script_output{script=\"json\",output=\"foo\"} 42",
		"script_output{script=\"json\",output=\"bar\"} 2.71828",
	}

	samples := runScripts(config.Scripts)

	assertSamples(t, samples, expectedSamples)
}

func TestScriptFilter(t *testing.T) {
	t.Run("RequiredParameters", func(t *testing.T) {
		_, err := scriptFilter(config.Scripts, "", "")

		if err.Error() != "`name` or `pattern` required" {
			t.Errorf("Expected failure when supplying no parameters")
		}
	})

	t.Run("NameMatch", func(t *testing.T) {
		scripts, err := scriptFilter(config.Scripts, "success", "")

		if err != nil {
			t.Errorf("Unexpected: %s", err.Error())
		}

		if len(scripts) != 1 || scripts[0] != config.Scripts[0] {
			t.Errorf("Expected script not found")
		}
	})

	t.Run("PatternMatch", func(t *testing.T) {
		scripts, err := scriptFilter(config.Scripts, "", "fail.*")

		if err != nil {
			t.Errorf("Unexpected: %s", err.Error())
		}

		if len(scripts) != 1 || scripts[0] != config.Scripts[1] {
			t.Errorf("Expected script not found")
		}
	})

	t.Run("AllMatch", func(t *testing.T) {
		scripts, err := scriptFilter(config.Scripts, "success", ".*")

		if err != nil {
			t.Errorf("Unexpected: %s", err.Error())
		}

		if len(scripts) != len(config.Scripts) {
			t.Fatalf("Expected %d scripts, received %d", len(config.Scripts), len(scripts))
		}

		for i, script := range config.Scripts {
			if scripts[i] != script {
				t.Fatalf("Expected script not found")
			}
		}
	})
}

func assertSamples(t *testing.T, samples []Sample, expected []any) {
	asserters := []SampleAsserter{}

	for _, exp := range expected {
		switch exp := exp.(type) {
		case SampleAsserter:
			asserters = append(asserters, exp)
		case Sample:
			asserters = append(asserters, ExactMatchAsserter{&exp})
		case string:
			if s, err := parseSample(exp); err == nil {
				asserters = append(asserters, ExactMatchAsserter{s})
			} else {
				t.Errorf("Failed to parse expected sample: %s", exp)
			}
		default:
			t.Logf("Unsupported type %T of expected Sample.", exp)
		}
	}

	assertSampleAsserters(t, samples, asserters)
}

func assertSampleAsserters(t *testing.T, samples []Sample, asserters []SampleAsserter) {
	if len(samples) != len(asserters) {
		t.Errorf("Expected %d samples, got %d", len(asserters), len(samples))
	}

	for _, asserter := range asserters {
		asserterMatchedASample := false
		for i, sample := range samples {
			if asserter.Assert(t, &sample) {
				asserterMatchedASample = true
				samples = slices.Delete(samples, i, i+1)
				break
			}
		}
		if !asserterMatchedASample {
			t.Errorf("Asserter %s did not match any samples.", asserter)
		}
	}
	for _, sample := range samples {
		t.Errorf("Unexpected sample %s was not matched by any asserter.", sample.String())
	}

}

func parseSample(s string) (sample *Sample, err error) {
	samplePattern := regexp.MustCompile(`^([-_a-zA-Z0-9]+)\s*[{]([^]]+)[}]\s+(\S+)$`)
	sampleParts := samplePattern.FindStringSubmatch(s)
	if sampleParts == nil {
		return nil, errors.New("Sample '" + s + "' does not match expected pattern " + samplePattern.String() + ".")
	}

	name := sampleParts[1]
	labelsPart := sampleParts[2]
	valuePart := sampleParts[3]

	value, err := strconv.ParseFloat(valuePart, 64)
	if err != nil {
		return nil, err
	}

	sample = &Sample{
		Name:   name,
		Labels: map[string]string{},
		Value:  value,
	}

	labels := strings.Split(labelsPart, ",")
	for _, label := range labels {
		labelParts := strings.SplitN(label, "=", 2)
		if len(labelParts) < 2 {
			return nil, errors.New("Sample contains label '" + label + "', which does not match expected form 'name=\"value\"'.")
		}
		labelName := strings.TrimSpace(labelParts[0])
		labelValue := strings.TrimSpace(labelParts[1])

		quotedStringPattern := regexp.MustCompile(`^"(.*)"$`)
		if q := quotedStringPattern.FindStringSubmatch(labelValue); q != nil {
			labelValue = q[1]
		}
		sample.Labels[labelName] = labelValue
	}

	return
}
