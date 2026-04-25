package main

import (
	"bytes"
	"math"
	"testing"
)

type OutputHandlerTestConfig struct {
	Name            string
	Output          string
	ProcessedOutput any
	Samples         []string
}

func TestNumberOutputHandler(t *testing.T) {
	handler := NumberOutputHandler{}

	testConfigs := []OutputHandlerTestConfig{

		{
			Name:            "text",
			Output:          "leet",
			ProcessedOutput: nil,
			Samples:         nil,
		},

		{
			Name:            "two_numbers",
			Output:          "19  79",
			ProcessedOutput: nil,
			Samples:         nil,
		},

		{
			Name:            "number_with_text",
			Output:          "10000000 dollars",
			ProcessedOutput: nil,
			Samples:         nil,
		},

		{
			Name:            "integer",
			Output:          "1337",
			ProcessedOutput: any(1337.0),
			Samples: []string{
				"script_output{script=\"integer\"} 1337.000000",
			},
		},

		{
			Name:            "positive_integer",
			Output:          "+1999",
			ProcessedOutput: any(1999.0),
			Samples: []string{
				"script_output{script=\"positive_integer\"} 1999.000000",
			},
		},

		{
			Name:            "negative_integer",
			Output:          "-1999",
			ProcessedOutput: any(-1999.0),
			Samples: []string{
				"script_output{script=\"negative_integer\"} -1999.000000",
			},
		},

		{
			Name:            "decimal",
			Output:          "23.42",
			ProcessedOutput: any(23.42),
			Samples: []string{
				"script_output{script=\"decimal\"} 23.420000",
			},
		},

		{
			Name:            "positive_decimal",
			Output:          "+2.71",
			ProcessedOutput: any(2.71),
			Samples: []string{
				"script_output{script=\"positive_decimal\"} 2.710000",
			},
		},

		{
			Name:            "negative_decimal",
			Output:          "-3.14",
			ProcessedOutput: any(-3.14),
			Samples: []string{
				"script_output{script=\"negative_decimal\"} -3.140000",
			},
		},

		{
			Name:            "number_with_padding",
			Output:          "  69  ",
			ProcessedOutput: any(69.0),
			Samples: []string{
				"script_output{script=\"number_with_padding\"} 69.000000",
			},
		},

		{
			Name:            "hex",
			Output:          "0x0000ff",
			ProcessedOutput: nil,
			Samples:         nil,
		},

		{
			Name:            "decimal_with_leading_zero",
			Output:          "0755",
			ProcessedOutput: any(755.0),
			Samples: []string{
				"script_output{script=\"decimal_with_leading_zero\"} 755.000000",
			},
		},

		{
			Name:            "not_a_number",
			Output:          "NaN",
			ProcessedOutput: any(math.NaN()),
			Samples: []string{
				"script_output{script=\"not_a_number\"} NaN",
			},
		},

		{
			Name:            "inf",
			Output:          "inf",
			ProcessedOutput: any(math.Inf(1)),
			Samples: []string{
				"script_output{script=\"inf\"} +Inf",
			},
		},

		{
			Name:            "infinity",
			Output:          "InfInIty",
			ProcessedOutput: any(math.Inf(1)),
			Samples: []string{
				"script_output{script=\"infinity\"} +Inf",
			},
		},

		{
			Name:            "positive_infinity",
			Output:          "+infinity",
			ProcessedOutput: any(math.Inf(1)),
			Samples: []string{
				"script_output{script=\"positive_infinity\"} +Inf",
			},
		},

		{
			Name:            "negative_infinity",
			Output:          "-iNfiNiTy",
			ProcessedOutput: any(math.Inf(-1)),
			Samples: []string{
				"script_output{script=\"negative_infinity\"} -Inf",
			},
		},
	}

	testHandler(t, handler, testConfigs)
}

func TestJsonOutputHandler(t *testing.T) {
	handler := JsonOutputHandler{}

	testConfigs := []OutputHandlerTestConfig{

		{
			Name:            "invalid_json",
			Output:          "{ not a mapping }",
			ProcessedOutput: nil,
			Samples:         nil,
		},

		{
			Name:            "true",
			Output:          "true",
			ProcessedOutput: any(true),
			Samples: []string{
				"script_output{script=\"true\",output=\".\"} 1.000000",
			},
		},

		{
			Name:            "false",
			Output:          "false",
			ProcessedOutput: any(false),
			Samples: []string{
				"script_output{script=\"false\",output=\".\"} 0.000000",
			},
		},

		{
			Name:            "number",
			Output:          "1701",
			ProcessedOutput: any(1701.0),
			Samples: []string{
				"script_output{script=\"number\",output=\".\"} 1701.000000",
			},
		},

		{
			Name:            "string",
			Output:          "\"howdy\"",
			ProcessedOutput: any("howdy"),
			Samples:         []string{},
		},

		{
			Name:            "numeric_string",
			Output:          "\"2001\"",
			ProcessedOutput: any("2001"),
			Samples: []string{
				"script_output{script=\"numeric_string\",output=\".\"} 2001.000000",
			},
		},

		{
			Name:            "array",
			Output:          "[1, 2, 4, 8, 16]",
			ProcessedOutput: any([]any{1.0, 2.0, 4.0, 8.0, 16.0}),
			Samples: []string{
				"script_output{script=\"array\",output=\"0\"} 1.000000",
				"script_output{script=\"array\",output=\"1\"} 2.000000",
				"script_output{script=\"array\",output=\"2\"} 4.000000",
				"script_output{script=\"array\",output=\"3\"} 8.000000",
				"script_output{script=\"array\",output=\"4\"} 16.000000",
			},
		},

		{
			Name:            "mixed_array",
			Output:          "[8000, null, \"42\", -0.0, \"ahoj!\", true, -3.14]",
			ProcessedOutput: any([]any{8000.0, nil, "42", 0.0, "ahoj!", true, -3.14}),
			Samples: []string{
				"script_output{script=\"mixed_array\",output=\"0\"} 8000.000000",
				"script_output{script=\"mixed_array\",output=\"2\"} 42.000000",
				"script_output{script=\"mixed_array\",output=\"3\"} -0.000000",
				"script_output{script=\"mixed_array\",output=\"5\"} 1.000000",
				"script_output{script=\"mixed_array\",output=\"6\"} -3.140000",
			},
		},

		{
			Name: "object",
			Output: "{" +
				"\"foo\": 42," +
				"\"bar\": 2.71828" +
				"}",
			ProcessedOutput: any(map[string]any{
				"foo": 42.0,
				"bar": 2.71828,
			}),
			Samples: []string{
				"script_output{script=\"object\",output=\"foo\"} 42.000000",
				"script_output{script=\"object\",output=\"bar\"} 2.718280",
			},
		},

		{
			Name: "mixed_object",
			Output: "{" +
				"\"text\": \"foo\"," +
				"\"number\": 7" +
				"}",
			ProcessedOutput: any(map[string]any{
				"text":   "foo",
				"number": 7.0,
			}),
			Samples: []string{
				"script_output{script=\"mixed_object\",output=\"number\"} 7.000000",
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
			ProcessedOutput: any(map[string]any{
				"text":    "foo",
				"number":  7.0,
				"boolean": true,
				"array": []any{
					true,
					2.0,
					"3",
				},
				"empty": []any{},
				"nested": any(map[string]any{
					"null":    nil,
					"boolean": false,
					"pi":      "3.14",
					"empty":   []any{},
				}),
			}),
			Samples: []string{
				"script_output{script=\"nested_json\",output=\"number\"} 7.000000",
				"script_output{script=\"nested_json\",output=\"boolean\"} 1.000000",
				"script_output{script=\"nested_json\",output=\"array.0\"} 1.000000",
				"script_output{script=\"nested_json\",output=\"array.1\"} 2.000000",
				"script_output{script=\"nested_json\",output=\"array.2\"} 3.000000",
				"script_output{script=\"nested_json\",output=\"nested.boolean\"} 0.000000",
				"script_output{script=\"nested_json\",output=\"nested.pi\"} 3.140000",
			},
		},
	}

	testHandler(t, handler, testConfigs)
}

func testHandler(t *testing.T, handler OutputHandler, testConfigs []OutputHandlerTestConfig) {

	for _, testConfig := range testConfigs {
		t.Run(
			"with_"+testConfig.Name,
			func(t *testing.T) {
				samples := handler.Handle(testConfig.Name, bytes.NewBufferString(testConfig.Output))
				assertEqualLinesInArbitraryOrder(t, samples, testConfig.Samples)
			},
		)
	}

}
