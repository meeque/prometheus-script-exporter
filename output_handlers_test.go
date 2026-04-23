package main

import (
	"bytes"
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"
)

type OutputHandlerTestConfig struct {
	Name 				    string
	TestOutput              string
	ExpectedProcessedOutput any
	ExpectedPrintedResult   string
}

func TestNumberOutputHandler(t *testing.T) {
	handler := NumberOutputHandler{}

	testConfigs := []OutputHandlerTestConfig{

		{
			Name:                   "text",
			TestOutput:             "leet",
			ExpectedProcessedOutput: nil,
			ExpectedPrintedResult:   "",
		},

		{
			Name:                    "two_numbers",
			TestOutput:              "19  79",
			ExpectedProcessedOutput: nil,
			ExpectedPrintedResult:   "",
		},

		{
			Name:                    "number_with_text",
			TestOutput:              "10000000 dollars",
			ExpectedProcessedOutput: nil,
			ExpectedPrintedResult:   "",
		},

		{
			Name:                    "integer",
			TestOutput:              "1337",
			ExpectedProcessedOutput: any(1337.0),
			ExpectedPrintedResult:   "script_output{script=\"integer\"} 1337.000000\n",
		},

		{
			Name:                    "positive_integer",
			TestOutput:              "+1999",
			ExpectedProcessedOutput: any(1999.0),
			ExpectedPrintedResult:   "script_output{script=\"positive_integer\"} 1999.000000\n",
		},

		{
			Name:                    "negative_integer",
			TestOutput:              "-1999",
			ExpectedProcessedOutput: any(-1999.0),
			ExpectedPrintedResult:   "script_output{script=\"negative_integer\"} -1999.000000\n",
		},

		{
			Name:                    "decimal",
			TestOutput:              "23.42",
			ExpectedProcessedOutput: any(23.42),
			ExpectedPrintedResult:   "script_output{script=\"decimal\"} 23.420000\n",
		},

		{
			Name:                    "positive_decimal",
			TestOutput:              "+2.71",
			ExpectedProcessedOutput: any(2.71),
			ExpectedPrintedResult:   "script_output{script=\"positive_decimal\"} 2.710000\n",
		},

		{
			Name:                    "negative_decimal",
			TestOutput:              "-3.14",
			ExpectedProcessedOutput: any(-3.14),
			ExpectedPrintedResult:   "script_output{script=\"negative_decimal\"} -3.140000\n",
		},

		{
			Name:                    "number_with_padding",
			TestOutput:              "  69  ",
			ExpectedProcessedOutput: any(69.0),
			ExpectedPrintedResult:   "script_output{script=\"number_with_padding\"} 69.000000\n",
		},

		{
			Name:                    "hex",
			TestOutput:              "0x0000ff",
			ExpectedProcessedOutput: nil,
			ExpectedPrintedResult:   "",
		},

		{
			Name:                    "decimal_with_leading_zero",
			TestOutput:              "0755",
			ExpectedProcessedOutput: any(755.0),
			ExpectedPrintedResult:   "script_output{script=\"decimal_with_leading_zero\"} 755.000000\n",
		},

		{
			Name:                    "not_a_number",
			TestOutput:              "NaN",
			ExpectedProcessedOutput: any(math.NaN()),
			ExpectedPrintedResult:   "script_output{script=\"not_a_number\"} NaN\n",
		},

		{
			Name:                    "inf",
			TestOutput:              "inf",
			ExpectedProcessedOutput: any(math.Inf(1)),
			ExpectedPrintedResult:   "script_output{script=\"inf\"} +Inf\n",
		},

		{
			Name:                    "infinity",
			TestOutput:              "InfInIty",
			ExpectedProcessedOutput: any(math.Inf(1)),
			ExpectedPrintedResult:   "script_output{script=\"infinity\"} +Inf\n",
		},

		{
			Name:                    "positive_infinity",
			TestOutput:              "+infinity",
			ExpectedProcessedOutput: any(math.Inf(1)),
			ExpectedPrintedResult:   "script_output{script=\"positive_infinity\"} +Inf\n",
		},

		{
			Name:                    "negative_infinity",
			TestOutput:              "-iNfiNiTy",
			ExpectedProcessedOutput: any(math.Inf(-1)),
			ExpectedPrintedResult:   "script_output{script=\"negative_infinity\"} -Inf\n",
		},

	}

	testHandlers(t, handler, testConfigs)
}

func TestJsonOutputHandler(t *testing.T) {
	handler := JsonOutputHandler{}

	testConfigs := []OutputHandlerTestConfig{

		{
			Name:                    "invalid_json",
			TestOutput:              "{ not a mapping }",
			ExpectedProcessedOutput: nil,
			ExpectedPrintedResult:   "",
		},

		{
			Name:                    "true",
			TestOutput:              "true",
			ExpectedProcessedOutput: any(true),
			ExpectedPrintedResult:   "script_output{script=\"true\",output=\".\"} 1.000000\n",
		},

		{
			Name:                    "false",
			TestOutput:              "false",
			ExpectedProcessedOutput: any(false),
			ExpectedPrintedResult:   "script_output{script=\"false\",output=\".\"} 0.000000\n",
		},

		{
			Name:                    "number",
			TestOutput:              "1701",
			ExpectedProcessedOutput: any(1701.0),
			ExpectedPrintedResult:   "script_output{script=\"number\",output=\".\"} 1701.000000\n",
		},

		{
			Name:                    "string",
			TestOutput:              "\"howdy\"",
			ExpectedProcessedOutput: any("howdy"),
			ExpectedPrintedResult:   "",
		},

		{
			Name:                    "numeric_string",
			TestOutput:              "\"2001\"",
			ExpectedProcessedOutput: any("2001"),
			ExpectedPrintedResult:   "script_output{script=\"numeric_string\",output=\".\"} 2001.000000\n",
		},

		{
			Name:                    "array",
			TestOutput:              "[1, 2, 4, 8, 16]",
			ExpectedProcessedOutput: any([]any{1.0, 2.0, 4.0, 8.0, 16.0}),
			ExpectedPrintedResult:   "script_output{script=\"array\",output=\"0\"} 1.000000\n" +
									 "script_output{script=\"array\",output=\"1\"} 2.000000\n" +
									 "script_output{script=\"array\",output=\"2\"} 4.000000\n" +
									 "script_output{script=\"array\",output=\"3\"} 8.000000\n" +
									 "script_output{script=\"array\",output=\"4\"} 16.000000\n",
		},

		{
			Name:                    "mixed_array",
			TestOutput:              "[8000, null, \"42\", -0.0, \"ahoj!\", true, -3.14]",
			ExpectedProcessedOutput: any([]any{8000.0, nil, "42", 0.0, "ahoj!", true, -3.14}),
			ExpectedPrintedResult:   "script_output{script=\"mixed_array\",output=\"0\"} 8000.000000\n" +
									 "script_output{script=\"mixed_array\",output=\"2\"} 42.000000\n" +
									 "script_output{script=\"mixed_array\",output=\"3\"} -0.000000\n" +
									 "script_output{script=\"mixed_array\",output=\"5\"} 1.000000\n" +
									 "script_output{script=\"mixed_array\",output=\"6\"} -3.140000\n",
		},

		{
			Name:                    "object",
			TestOutput:              "{" +
									   "\"foo\": 42," +
									   "\"bar\": 2.71828" +
									 "}",
			ExpectedProcessedOutput: any(map[string]any{
									   "foo": 42.0,
									   "bar": 2.71828,
									 }),
			ExpectedPrintedResult:   "script_output{script=\"object\",output=\"foo\"} 42.000000\n" +
									 "script_output{script=\"object\",output=\"bar\"} 2.718280\n",
		},

		{
			Name:                    "mixed_object",
			TestOutput:              "{" +
									   "\"text\": \"foo\"," +
									   "\"number\": 7" +
									 "}",
			ExpectedProcessedOutput: any(map[string]any{
			                           "text": "foo",
									   "number": 7.0,
									 }),
			ExpectedPrintedResult:   "script_output{script=\"mixed_object\",output=\"number\"} 7.000000\n",
		},

		{
			Name:                    "nested_json",
			TestOutput:              "{" +
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
			ExpectedProcessedOutput: any(map[string]any{
									   "text": "foo",
									   "number": 7.0,
									   "boolean": true,
									   "array": []any{
										  true,
										  2.0,
										  "3",
									   },
									   "empty": []any{},
									   "nested": any(map[string]any{
										  "null": nil,
										  "boolean": false,
										  "pi": "3.14",
										  "empty": []any{},
									   }),
									 }),
			ExpectedPrintedResult:   "script_output{script=\"nested_json\",output=\"number\"} 7.000000\n" +
									 "script_output{script=\"nested_json\",output=\"boolean\"} 1.000000\n" +
									 "script_output{script=\"nested_json\",output=\"array.0\"} 1.000000\n" +
									 "script_output{script=\"nested_json\",output=\"array.1\"} 2.000000\n" +
									 "script_output{script=\"nested_json\",output=\"array.2\"} 3.000000\n" +
									 "script_output{script=\"nested_json\",output=\"nested.boolean\"} 0.000000\n" +
									 "script_output{script=\"nested_json\",output=\"nested.pi\"} 3.140000\n",
		},

	}

	testHandlers(t, handler, testConfigs)
}

func testHandlers(t *testing.T, handler OutputHandler, testConfigs []OutputHandlerTestConfig) {
	for _, testConfig := range testConfigs {
		t.Run(
			fmt.Sprintf("with_%s", testConfig.Name),
			func(t *testing.T) {
				testHandler(t, handler, &testConfig)
			},
		)
	}
}

func testHandler(t *testing.T, handler OutputHandler, testConfig *OutputHandlerTestConfig) {
	processedOutput, err := handler.Process(bytes.NewBufferString(testConfig.TestOutput))

	if err != nil {
		if testConfig.ExpectedProcessedOutput == nil {
			return
		} else {
			t.Errorf("Unexpected error when processing output: %s", err)
		}
	} else {
		if testConfig.ExpectedProcessedOutput == nil {
			t.Errorf("Expected an error when processing output, but got none")
		}
	}

	if !deepEqualPointers(&processedOutput, &testConfig.ExpectedProcessedOutput) {
		t.Errorf("Expected output %s != %s", processedOutput, testConfig.ExpectedProcessedOutput)
	}

	printWriter := &bytes.Buffer{}
	handler.Print(printWriter, testConfig.Name, processedOutput)
	printedResult := printWriter.String()

	if !equalLinesInArbitraryOrder(printedResult, testConfig.ExpectedPrintedResult) {
		t.Errorf("Expected printed result '%s' != '%s'", printedResult, testConfig.ExpectedPrintedResult)
	}
}

func equalLinesInArbitraryOrder(string1 string, string2 string) bool {
	lines1 := strings.Split(string1, "\n")[:]
	lines2 := strings.Split(string2, "\n")[:]

	slices.Sort(lines1)
	slices.Sort(lines2)

	return slices.Equal(lines1, lines2)
}
