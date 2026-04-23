package main

import (
	"fmt"
	"reflect"
	"testing"
)

type ExpectedMeasurement struct {
	Success     int
	Status      int
	MinDuration float64
	Output      *any
}

type ExpectedMeasurements map[string]ExpectedMeasurement

var config = &Config{
	Scripts: []*Script{
		{"success", "exit 0", 15, ""},
		{"failure", "exit 1", 15, ""},
		{"timeout", "sleep 3", 1, ""},
		{"number", "echo 23", 15, "number"},
		{"invalid-num", "echo twentythree", 15, "number"},
		{"json", "echo '{\"foo\": 42, \"bar\": 2.71828}'", 15, "json"},
		{"invalid-json", "echo '{ not a mapping }'", 15, "json"},
		{"mixed-json", "echo '{\"text\": \"foo\", \"number\": 7}'", 15, "json"},
	},
}

func TestRunScripts(t *testing.T) {
	measurements := runScripts(config.Scripts)

	twentyThree := any(23.0)

	fooBarMap := any(
		map[string]any{
			"foo": 42.0,
			"bar": 2.71828,
		})

	textNumberMap := any(
		map[string]any{
			"text": "foo",
			"number": 7.0,
		})

	expectedMeasurements := ExpectedMeasurements{
		"success": {1, 0, 0, nil},
		"failure": {0, 1, 0, nil},
		"timeout": {0, -1, 0.9, nil},
		"number":  {1, 0, 0, &twentyThree},
		"invalid-num": {1, 0, 0, nil},
		"json":    {1, 0, 0, &fooBarMap},
		"invalid-json": {1, 0, 0, nil},
		"mixed-json": {1, 0, 0, &textNumberMap},
	}

	if len(measurements) != len(config.Scripts) {
		t.Errorf("Expected %d measurements, received %d", len(config.Scripts), len(measurements))
	}

	for _, measurement := range measurements {
		expectedResult, ok := expectedMeasurements[measurement.Script.Name]

		if !ok {
			t.Errorf("Got a measurement for an unexpected script: %s", measurement.Script.Name)
			continue
		}

		if measurement.Success != expectedResult.Success {
			t.Errorf("Expected success %d != %d: %s", measurement.Success, expectedResult.Success, measurement.Script.Name)
		}

		if measurement.Status != expectedResult.Status {
			t.Errorf("Expected status %d != %d: %s", measurement.Status, expectedResult.Status, measurement.Script.Name)
		}

		if measurement.Duration < expectedResult.MinDuration {
			t.Errorf("Expected duration %f < %f: %s", measurement.Duration, expectedResult.MinDuration, measurement.Script.Name)
		}

		if !deepEqualPointers(measurement.Output, expectedResult.Output) {
			t.Errorf("Expected output %s != %s: %s", stringPointer(measurement.Output), stringPointer(expectedResult.Output), measurement.Script.Name)
		}
	}
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
	return reflect.DeepEqual(*a, *b)
}

func stringPointer(p *any) string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%v", *p)
}
