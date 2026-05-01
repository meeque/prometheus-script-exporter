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

	samples := runScripts(scriptExporterTestConfig.Scripts)

	assertSamples(t, samples, &expectedSamples)
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
