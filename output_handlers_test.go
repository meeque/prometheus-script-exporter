package main

import (
	"bytes"
	"fmt"
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

	if err == nil {
		if testConfig.ExpectedProcessedOutput == nil {
			t.Errorf("Expected an error when processing output, but got none")
		}
	} else {
		if testConfig.ExpectedProcessedOutput != nil {
			t.Errorf("Unexpected error when processing output")
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
