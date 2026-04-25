package main

import (
	"regexp"
	"slices"
	"strconv"
	"testing"
)

type SampleAsserter interface {
	Assert(t *testing.T, sample string) (isMatch bool)
}

type MinDurationAsserter struct {
	SamplePattern   regexp.Regexp
	DurationPattern regexp.Regexp
	Min             float64
}

func (a MinDurationAsserter) Assert(t *testing.T, sample string) (isMatch bool) {
	isMatch = a.SamplePattern.Match([]byte(sample))
	if isMatch {
		durationString := a.DurationPattern.Find([]byte(sample))
		if durationString == nil {
			t.Errorf("Could not find duration pattern %s in sample %s", a.DurationPattern.String(), sample)
			return

		}
		duration, err := strconv.ParseFloat(string(durationString), 64)
		if err != nil {
			t.Errorf("Could not parse sampled duration %s as number: %s", string(durationString), err)
			return
		}
		if duration < a.Min {
			t.Errorf("Expected sampled duration to be at least %f, but got %f", a.Min, duration)
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
	samples := runScripts(config.Scripts)

	expectedSamples := []string{
		"script_duration_seconds{script=\"success\"} ?.??",
		"script_status{script=\"success\"} 0",
		"script_success{script=\"success\"} 1",

		"script_duration_seconds{script=\"failure\"} ?.??",
		"script_status{script=\"failure\"} 1",
		"script_success{script=\"failure\"} 0",

		"script_duration_seconds{script=\"timeout\"} ?.??",
		"script_status{script=\"timeout\"} -1",
		"script_success{script=\"timeout\"} 0",

		"script_duration_seconds{script=\"number\"} ?.??",
		"script_status{script=\"number\"} 0",
		"script_success{script=\"number\"} 1",
		"script_output{script=\"number\"} 23.000000",

		"script_duration_seconds{script=\"json\"} 0.001845",
		"script_status{script=\"json\"} 0",
		"script_success{script=\"json\"} 1",
		"script_output{script=\"json\",output=\"foo\"} 42.000000",
		"script_output{script=\"json\",output=\"bar\"} 2.718280",
	}

	valueRegexp := regexp.MustCompile(`[^\s]+$`)

	sampleAsserters := []SampleAsserter{
		MinDurationAsserter{
			*regexp.MustCompile(`^script_duration_seconds[{].*script="success".*[}]\s+`),
			*valueRegexp,
			0.0,
		},
		MinDurationAsserter{
			*regexp.MustCompile(`^script_duration_seconds[{].*script="failure".*[}]\s+`),
			*valueRegexp,
			0.0,
		},
		MinDurationAsserter{
			*regexp.MustCompile(`^script_duration_seconds[{].*script="timeout".*[}]\s+`),
			*valueRegexp,
			0.9,
		},
		MinDurationAsserter{
			*regexp.MustCompile(`^script_duration_seconds[{].*script="number".*[}]\s+`),
			*valueRegexp,
			0.0,
		},
		MinDurationAsserter{
			*regexp.MustCompile(`^script_duration_seconds[{].*script="json".*[}]\s+`),
			*valueRegexp,
			0.0,
		},
	}

	assertEqualLinesInArbitraryOrder(t, samples, expectedSamples, sampleAsserters)
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

func assertEqualLinesInArbitraryOrder(t *testing.T, actual []string, expected []string, asserters []SampleAsserter) {
	if len(actual) != len(expected) {
		t.Errorf("Expected %d lines, got %d", len(expected), len(actual))
	}

	slices.Sort(actual)
	slices.Sort(expected)

	for i, exp := range expected {
		act := actual[i]
		asserterMatch := false
		for _, asserter := range asserters {
			asserterMatch = asserter.Assert(t, act)
			if asserterMatch {
				break
			}
		}
		if asserterMatch {
			continue
		}
		if act != exp {
			t.Errorf("Expected line %d to be %q, got %q", i, exp, act)
		}
	}
}
