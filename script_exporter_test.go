package main

import (
	"math"
	"reflect"
	"slices"
	"testing"
)

type ExpectedMeasurement struct {
	Success     int
	Status      int
	MinDuration float64
	SampleCount int
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
		"script_duration_seconds{script=\"success\"} 0.001697",
		"script_status{script=\"success\"} 0",
		"script_success{script=\"success\"} 1",

		"script_duration_seconds{script=\"failure\"} 0.001439",
		"script_status{script=\"failure\"} 1",
		"script_success{script=\"failure\"} 0",

		"script_duration_seconds{script=\"timeout\"} 1.001120",
		"script_status{script=\"timeout\"} -1",
		"script_success{script=\"timeout\"} 0",

		"script_duration_seconds{script=\"number\"} 0.002202",
		"script_status{script=\"number\"} 0",
		"script_success{script=\"number\"} 1",
		"script_output{script=\"number\"} 23.000000",

		"script_duration_seconds{script=\"json\"} 0.001845",
		"script_status{script=\"json\"} 0",
		"script_success{script=\"json\"} 1",
		"script_output{script=\"json\",output=\"foo\"} 42.000000",
		"script_output{script=\"json\",output=\"bar\"} 2.718280",
	}

	assertEqualLinesInArbitraryOrder(t, samples, expectedSamples)
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

func deepEqualPointers(a, b *any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	aAsFloat, aIsFloat := (*a).(float64)
	bAsFloat, bIsFloat := (*b).(float64)
	if aIsFloat && bIsFloat && math.IsNaN(aAsFloat) && math.IsNaN(bAsFloat) {
		return true
	}

	return reflect.DeepEqual(*a, *b)
}

func assertEqualLinesInArbitraryOrder(t *testing.T, actual []string, expected []string) {
	if len(actual) != len(expected) {
		t.Errorf("Expected %d lines, got %d", len(expected), len(actual))
	}

	slices.Sort(actual)
	slices.Sort(expected)

	for i, exp := range expected {
		act := actual[i]
		if act != exp {
			t.Errorf("Expected line %d to be %q, got %q", i, exp, act)
		}
	}
}
