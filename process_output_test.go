package main

import (
	"bytes"
	"math"
	"testing"
)

type ProcessOutputTestConfig struct {
	Name    string
	Output  string
	Samples []Sample
}

func TestProcessNumberOutput(t *testing.T) {

	testConfigs := []ProcessOutputTestConfig{

		{
			Name:    "text",
			Output:  "leet",
			Samples: nil,
		},

		{
			Name:    "two_numbers",
			Output:  "19  79",
			Samples: nil,
		},

		{
			Name:    "number_with_text",
			Output:  "10000000 dollars",
			Samples: nil,
		},

		{
			Name:   "integer",
			Output: "1337",
			Samples: []Sample{
				*NewNumberOutputSample("integer", 1337.0),
			},
		},

		{
			Name:   "positive_integer",
			Output: "+1999",
			Samples: []Sample{
				*NewNumberOutputSample("positive_integer", 1999),
			},
		},

		{
			Name:   "negative_integer",
			Output: "-1999",
			Samples: []Sample{
				*NewNumberOutputSample("negative_integer", -1999),
			},
		},

		{
			Name:   "decimal",
			Output: "23.42",
			Samples: []Sample{
				*NewNumberOutputSample("decimal", 23.42),
			},
		},

		{
			Name:   "positive_decimal",
			Output: "+2.71",
			Samples: []Sample{
				*NewNumberOutputSample("positive_decimal", 2.71),
			},
		},

		{
			Name:   "negative_decimal",
			Output: "-3.14",
			Samples: []Sample{
				*NewNumberOutputSample("negative_decimal", -3.14),
			},
		},

		{
			Name:   "number_with_padding",
			Output: "  69  ",
			Samples: []Sample{
				*NewNumberOutputSample("number_with_padding", 69),
			},
		},

		{
			Name:    "hex",
			Output:  "0x0000ff",
			Samples: nil,
		},

		{
			Name:   "decimal_with_leading_zero",
			Output: "0755",
			Samples: []Sample{
				*NewNumberOutputSample("decimal_with_leading_zero", 755),
			},
		},

		{
			Name:   "not_a_number",
			Output: "NaN",
			Samples: []Sample{
				*NewNumberOutputSample("not_a_number", math.NaN()),
			},
		},

		{
			Name:   "inf",
			Output: "inf",
			Samples: []Sample{
				*NewNumberOutputSample("inf", math.Inf(1)),
			},
		},

		{
			Name:   "infinity",
			Output: "InfInIty",
			Samples: []Sample{
				*NewNumberOutputSample("infinity", math.Inf(1)),
			},
		},

		{
			Name:   "positive_infinity",
			Output: "+infinity",
			Samples: []Sample{
				*NewNumberOutputSample("positive_infinity", math.Inf(1)),
			},
		},

		{
			Name:   "negative_infinity",
			Output: "-iNfiNiTy",
			Samples: []Sample{
				*NewNumberOutputSample("negative_infinity", math.Inf(-1)),
			},
		},
	}

	testProcessOutput(t, ProcessNumberOutput, testConfigs)
}

func TestProcessJsonOutput(t *testing.T) {

	testConfigs := []ProcessOutputTestConfig{

		{
			Name:    "invalid_json",
			Output:  "{ not a mapping }",
			Samples: nil,
		},

		{
			Name:   "true",
			Output: "true",
			Samples: []Sample{
				*NewJsonOutputSample("true", ".", 1),
			},
		},

		{
			Name:   "false",
			Output: "false",
			Samples: []Sample{
				*NewJsonOutputSample("false", ".", 0),
			},
		},

		{
			Name:   "number",
			Output: "1701",
			Samples: []Sample{
				*NewJsonOutputSample("number", ".", 1701),
			},
		},

		{
			Name:    "string",
			Output:  "\"howdy\"",
			Samples: nil,
		},

		{
			Name:   "numeric_string",
			Output: "\"2001\"",
			Samples: []Sample{
				*NewJsonOutputSample("numeric_string", ".", 2001),
			},
		},

		{
			Name:   "array",
			Output: "[1, 2, 4, 8, 16]",
			Samples: []Sample{
				*NewJsonOutputSample("array", "0", 1),
				*NewJsonOutputSample("array", "1", 2),
				*NewJsonOutputSample("array", "2", 4),
				*NewJsonOutputSample("array", "3", 8),
				*NewJsonOutputSample("array", "4", 16),
			},
		},

		{
			Name:   "mixed_array",
			Output: "[8000, null, \"42\", -0.0, \"ahoj!\", true, -3.14]",
			Samples: []Sample{
				*NewJsonOutputSample("mixed_array", "0", 8000),
				*NewJsonOutputSample("mixed_array", "2", 42),
				*NewJsonOutputSample("mixed_array", "3", math.Copysign(0, -1)),
				*NewJsonOutputSample("mixed_array", "5", 1),
				*NewJsonOutputSample("mixed_array", "6", -3.14),
			},
		},

		{
			Name: "object",
			Output: "{" +
				"\"foo\": 42," +
				"\"bar\": 2.71828" +
				"}",
			Samples: []Sample{
				*NewJsonOutputSample("object", "foo", 42),
				*NewJsonOutputSample("object", "bar", 2.71828),
			},
		},

		{
			Name: "mixed_object",
			Output: "{" +
				"\"text\": \"foo\"," +
				"\"number\": 7" +
				"}",
			Samples: []Sample{
				*NewJsonOutputSample("mixed_object", "number", 7),
			},
		},

		{
			Name: "nested_json",
			Output: "{" +
				"\"text\": \"foo\"," +
				"\"number\": 7," +
				"\"boolean\": true," +
				"\"array\": [true, 2, \"3\"]," +
				"\"empty\": []," +
				"\"nested\": {" +
				"\"null\": null," +
				"\"boolean\": false," +
				"\"pi\": \"3.14\"," +
				"\"empty\": []" +
				"}" +
				"}",
			Samples: []Sample{
				*NewJsonOutputSample("nested_json", "number", 7),
				*NewJsonOutputSample("nested_json", "boolean", 1),
				*NewJsonOutputSample("nested_json", "array.0", 1),
				*NewJsonOutputSample("nested_json", "array.1", 2),
				*NewJsonOutputSample("nested_json", "array.2", 3),
				*NewJsonOutputSample("nested_json", "nested.boolean", 0),
				*NewJsonOutputSample("nested_json", "nested.pi", 3.14),
			},
		},
	}

	testProcessOutput(t, ProcessJsonOutput, testConfigs)
}

func testProcessOutput(t *testing.T, processor ProcessOutput, testConfigs []ProcessOutputTestConfig) {
	for _, testConfig := range testConfigs {
		t.Run(
			"with_"+testConfig.Name,
			func(t *testing.T) {
				samples, err := processor(testConfig.Name, bytes.NewBufferString(testConfig.Output))
				if err != nil && testConfig.Samples == nil {
					return
				}

				expectedSamples := make([]any, len(testConfig.Samples))
				for i, s := range testConfig.Samples {
					expectedSamples[i] = s
				}
				assertSamples(t, samples, &expectedSamples)
			},
		)
	}
}
