package main

import (
	"testing"
)

var scriptExporterTestConfig = &Config{
	Scripts: []*Script{
		{"success", "exit 0", 15, ""},
		{"failure", "exit 1", 15, ""},
		{"timeout", "sleep 3", 1, ""},
		{"number", "echo 23", 15, "number"},
		{"json", "echo '{\"foo\": 42, \"bar\": 2.71828}'", 15, "json"},
	},
}

func TestRunScripts(t *testing.T) {
	asserters := []AssertSamples{
		NewAssertDurationSamples(t),
		NewAssertSamplesEqual(t),
	}

	expectedSamples := []Sample{
		*NewScriptSample("script_duration_seconds", "success", 0), // 0 - 0.5
		*NewScriptSample("script_status", "success", 0),
		*NewScriptSample("script_success", "success", 1),

		*NewScriptSample("script_duration_seconds", "failure", 0), // 0 - 0.5
		*NewScriptSample("script_status", "failure", 1),
		*NewScriptSample("script_success", "failure", 0),

		*NewScriptSample("script_duration_seconds", "timeout", 0), // 0.9 - 1.4
		*NewScriptSample("script_status", "timeout", -1),
		*NewScriptSample("script_success", "timeout", 0),

		*NewScriptSample("script_duration_seconds", "number", 0), // 0 - 0.5
		*NewScriptSample("script_status", "number", 0),
		*NewScriptSample("script_success", "number", 1),
		*NewNumberOutputSample("number", 23),

		*NewScriptSample("script_duration_seconds", "json", 0), // 0 - 0.5
		*NewScriptSample("script_status", "json", 0),
		*NewScriptSample("script_success", "json", 1),
		*NewJsonOutputSample("json", "foo", 42),
		*NewJsonOutputSample("json", "bar", 2.71828),
	}

	actualSamples := runScripts(scriptExporterTestConfig.Scripts)

	assertSamples(t, asserters, *actualSamples, expectedSamples)
}

func TestScriptFilter(t *testing.T) {
	t.Run("RequiredParameters", func(t *testing.T) {
		_, err := scriptFilter(scriptExporterTestConfig.Scripts, "", "")

		if err.Error() != "`name` or `pattern` required" {
			t.Errorf("Expected failure when supplying no parameters")
		}
	})

	t.Run("NameMatch", func(t *testing.T) {
		scripts, err := scriptFilter(scriptExporterTestConfig.Scripts, "success", "")

		if err != nil {
			t.Errorf("Unexpected: %s", err.Error())
		}

		if len(scripts) != 1 || scripts[0] != scriptExporterTestConfig.Scripts[0] {
			t.Errorf("Expected script not found")
		}
	})

	t.Run("PatternMatch", func(t *testing.T) {
		scripts, err := scriptFilter(scriptExporterTestConfig.Scripts, "", "fail.*")

		if err != nil {
			t.Errorf("Unexpected: %s", err.Error())
		}

		if len(scripts) != 1 || scripts[0] != scriptExporterTestConfig.Scripts[1] {
			t.Errorf("Expected script not found")
		}
	})

	t.Run("AllMatch", func(t *testing.T) {
		scripts, err := scriptFilter(scriptExporterTestConfig.Scripts, "success", ".*")

		if err != nil {
			t.Errorf("Unexpected: %s", err.Error())
		}

		if len(scripts) != len(scriptExporterTestConfig.Scripts) {
			t.Fatalf("Expected %d scripts, received %d", len(scriptExporterTestConfig.Scripts), len(scripts))
		}

		for i, script := range scriptExporterTestConfig.Scripts {
			if scripts[i] != script {
				t.Fatalf("Expected script not found")
			}
		}
	})
}

func NewAssertDurationSamples(t *testing.T) AssertSamples {
	return func(actual, expected Sample) (done bool) {
		if actual.Name != "script_duration_seconds" {
			return false
		}
		if !expected.EqualNameAndLabels(actual) {
			t.Errorf("Expected Sample '%s' but got '%s'", expected.StringNameAndLabels(), actual.StringNameAndLabels())
			return true
		}
		switch(actual.Labels["script"])  {
		case "timeout":
			if actual.Value < 0.9 || actual.Value > 1.4 {
				t.Errorf("Expected duration to be between %f and %f, but got %f", 0.9, 1.4, actual.Value)
			}
		default:
			if actual.Value < 0 || actual.Value > 0.5 {
				t.Errorf("Expected duration to be between %f and %f, but got %f", 0.0, 0.5, actual.Value)
			}
		}
		return true
	}
}
