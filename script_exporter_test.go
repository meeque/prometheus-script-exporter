package main

import (
	"maps"
	"math"
	"reflect"
	"slices"
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
		*NewScriptSample("script_status", "success", 0),
		*NewScriptSample("script_success", "success", 1),

		MinDurationAsserter{
			"script_duration_seconds",
			map[string]string{"script": "failure"},
			0.0,
			0.5,
		},
		*NewScriptSample("script_status", "failure", 1),
		*NewScriptSample("script_success", "failure", 0),

		MinDurationAsserter{
			"script_duration_seconds",
			map[string]string{"script": "timeout"},
			0.9,
			1.4,
		},
		*NewScriptSample("script_status", "timeout", -1),
		*NewScriptSample("script_success", "timeout", 0),

		MinDurationAsserter{
			"script_duration_seconds",
			map[string]string{"script": "number"},
			0.0,
			0.5,
		},
		*NewScriptSample("script_status", "number", 0),
		*NewScriptSample("script_success", "number", 1),
		*NewNumberOutputSample("number", 23),

		MinDurationAsserter{
			"script_duration_seconds",
			map[string]string{"script": "json"},
			0.0,
			0.5,
		},
		*NewScriptSample("script_status", "json", 0),
		*NewScriptSample("script_success", "json", 1),
		*NewJsonOutputSample("json", "foo", 42),
		*NewJsonOutputSample("json", "bar", 2.71828),
	}

	samples := runScripts(config.Scripts)

	assertSamples(t, samples, &expectedSamples)
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

func assertSamples(t *testing.T, samples *[]Sample, expected *[]any) {
	asserters := []SampleAsserter{}

	for _, exp := range *expected {
		switch exp := exp.(type) {
		case SampleAsserter:
			asserters = append(asserters, exp)
		case Sample:
			asserters = append(asserters, ExactMatchAsserter{&exp})
		default:
			t.Logf("Unsupported type %T of expected Sample.", exp)
		}
	}

	assertSampleAsserters(t, samples, asserters)
}

func assertSampleAsserters(t *testing.T, samples *[]Sample, asserters []SampleAsserter) {
	if samples == nil && asserters == nil {
		return
	}
	if len(*samples) != len(asserters) {
		t.Errorf("Expected %d samples, got %d", len(asserters), len(*samples))
	}

	for _, asserter := range asserters {
		asserterMatchedASample := false
		for i, sample := range *samples {
			if asserter.Assert(t, &sample) {
				asserterMatchedASample = true
				*samples = slices.Delete(*samples, i, i+1)
				break
			}
		}
		if !asserterMatchedASample {
			t.Errorf("Asserter %s did not match any samples.", asserter)
		}
	}
	for _, sample := range *samples {
		t.Errorf("Unexpected sample %s was not matched by any asserter.", sample.String())
	}

}
