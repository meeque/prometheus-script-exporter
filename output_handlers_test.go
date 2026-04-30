package main

import (
	"bytes"
	"testing"
)

type OutputHandlerTestConfig struct {
	Name    string
	Output  string
	Samples []string
}

func TestNumberOutputHandler(t *testing.T) {
	handler := NumberOutputHandler{}

	testConfigs := []OutputHandlerTestConfig{

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
			Samples: []string{
				"script_output{script=\"integer\"} 1337",
			},
		},

		{
			Name:   "positive_integer",
			Output: "+1999",
			Samples: []string{
				"script_output{script=\"positive_integer\"} 1999",
			},
		},

		{
			Name:   "negative_integer",
			Output: "-1999",
			Samples: []string{
				"script_output{script=\"negative_integer\"} -1999",
			},
		},

		{
			Name:   "decimal",
			Output: "23.42",
			Samples: []string{
				"script_output{script=\"decimal\"} 23.42",
			},
		},

		{
			Name:   "positive_decimal",
			Output: "+2.71",
			Samples: []string{
				"script_output{script=\"positive_decimal\"} 2.71",
			},
		},

		{
			Name:   "negative_decimal",
			Output: "-3.14",
			Samples: []string{
				"script_output{script=\"negative_decimal\"} -3.14",
			},
		},

		{
			Name:   "number_with_padding",
			Output: "  69  ",
			Samples: []string{
				"script_output{script=\"number_with_padding\"} 69",
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
			Samples: []string{
				"script_output{script=\"decimal_with_leading_zero\"} 755",
			},
		},

		{
			Name:   "not_a_number",
			Output: "NaN",
			Samples: []string{
				"script_output{script=\"not_a_number\"} NaN",
			},
		},

		{
			Name:   "inf",
			Output: "inf",
			Samples: []string{
				"script_output{script=\"inf\"} +Inf",
			},
		},

		{
			Name:   "infinity",
			Output: "InfInIty",
			Samples: []string{
				"script_output{script=\"infinity\"} +Inf",
			},
		},

		{
			Name:   "positive_infinity",
			Output: "+infinity",
			Samples: []string{
				"script_output{script=\"positive_infinity\"} +Inf",
			},
		},

		{
			Name:   "negative_infinity",
			Output: "-iNfiNiTy",
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
			Name:    "invalid_json",
			Output:  "{ not a mapping }",
			Samples: nil,
		},

		{
			Name:   "true",
			Output: "true",
			Samples: []string{
				"script_output{script=\"true\",output=\".\"} 1",
			},
		},

		{
			Name:   "false",
			Output: "false",
			Samples: []string{
				"script_output{script=\"false\",output=\".\"} 0",
			},
		},

		{
			Name:   "number",
			Output: "1701",
			Samples: []string{
				"script_output{script=\"number\",output=\".\"} 1701",
			},
		},

		{
			Name:    "string",
			Output:  "\"howdy\"",
			Samples: []string{},
		},

		{
			Name:   "numeric_string",
			Output: "\"2001\"",
			Samples: []string{
				"script_output{script=\"numeric_string\",output=\".\"} 2001",
			},
		},

		{
			Name:   "array",
			Output: "[1, 2, 4, 8, 16]",
			Samples: []string{
				"script_output{script=\"array\",output=\"0\"} 1",
				"script_output{script=\"array\",output=\"1\"} 2",
				"script_output{script=\"array\",output=\"2\"} 4",
				"script_output{script=\"array\",output=\"3\"} 8",
				"script_output{script=\"array\",output=\"4\"} 16",
			},
		},

		{
			Name:   "mixed_array",
			Output: "[8000, null, \"42\", -0.0, \"ahoj!\", true, -3.14]",
			Samples: []string{
				"script_output{script=\"mixed_array\",output=\"0\"} 8000",
				"script_output{script=\"mixed_array\",output=\"2\"} 42",
				"script_output{script=\"mixed_array\",output=\"3\"} -0",
				"script_output{script=\"mixed_array\",output=\"5\"} 1",
				"script_output{script=\"mixed_array\",output=\"6\"} -3.14",
			},
		},

		{
			Name: "object",
			Output: "{" +
				"\"foo\": 42," +
				"\"bar\": 2.71828" +
				"}",
			Samples: []string{
				"script_output{script=\"object\",output=\"foo\"} 42",
				"script_output{script=\"object\",output=\"bar\"} 2.71828",
			},
		},

		{
			Name: "mixed_object",
			Output: "{" +
				"\"text\": \"foo\"," +
				"\"number\": 7" +
				"}",
			Samples: []string{
				"script_output{script=\"mixed_object\",output=\"number\"} 7",
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
			Samples: []string{
				"script_output{script=\"nested_json\",output=\"number\"} 7",
				"script_output{script=\"nested_json\",output=\"boolean\"} 1",
				"script_output{script=\"nested_json\",output=\"array.0\"} 1",
				"script_output{script=\"nested_json\",output=\"array.1\"} 2",
				"script_output{script=\"nested_json\",output=\"array.2\"} 3",
				"script_output{script=\"nested_json\",output=\"nested.boolean\"} 0",
				"script_output{script=\"nested_json\",output=\"nested.pi\"} 3.14",
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
				expectedSamples := make([]any, len(testConfig.Samples))
				for i, s := range testConfig.Samples {
					expectedSamples[i] = s
				}
				assertSamples(t, &samples, expectedSamples)
			},
		)
	}

}
