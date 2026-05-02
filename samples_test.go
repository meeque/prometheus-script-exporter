package main

import (
	"math"
	"slices"
	"strings"
	"testing"
)

func TestSampleEqual(t *testing.T) {

	testSamples1 := []Sample{
		*NewSample("foo", map[string]string{}, 0),
		*NewSample("bar", map[string]string{"a": "A", "b": "B"}, -0.7),
		*NewScriptSample("test_metric", "test_script", math.NaN()),
	}

	testSamples2 := []Sample{
		*NewSample("foo", map[string]string{}, 0.0),
		*NewSample("bar", map[string]string{"b": "B", "a": "A"}, -0.7),
		*NewSample("test_metric", map[string]string{"script": "test_script"}, math.NaN()),
	}

	testSamplesEqual(t, testSamples1, testSamples2)
}

func TestSampleString(t *testing.T) {

	testSamplesAndExpectedStrings := map[*Sample]string{
		NewSample("foo", map[string]string{}, 256e-2): "foo{} 2.56",
		NewSample("bar", map[string]string{}, -0.7):   "bar{} -0.7",

		NewSample("labeled", map[string]string{"xxx": "yyy", "foo": "bar"}, 256):                                     `labeled{foo="bar",xxx="yyy"} 256`,
		NewSample("labeled_with_escaped_value", map[string]string{"foo": "quote\"backslash\\newline\n"}, math.NaN()): "labeled_with_escaped_value{foo=\"quote\\\"backslash\\\\newline\\n\"} NaN",
		NewSample("labeled_with_escaped_name", map[string]string{"\"\\\n": "bar"}, math.Inf(1)):                      "labeled_with_escaped_name{\"\\\"\\\\\\n\"=\"bar\"} +Inf",
		NewSample("labeled_with_escaped_name", map[string]string{"{a=b},c": "foo"}, math.Inf(-1)):                    "labeled_with_escaped_name{\"{a=b},c\"=\"foo\"} -Inf",

		NewScriptSample("sample_name", "script_name", -0): `sample_name{script="script_name"} 0`,
	}

	for testSample, expectedString := range testSamplesAndExpectedStrings {
		sampleString := testSample.String()
		if sampleString != expectedString {
			t.Errorf("Expected Sample.String() to be '%s' but got '%s'", expectedString, sampleString)
		}
	}
}

type AssertSamples func(t *testing.T, actual, expected Sample) (done bool)

func AssertSamplesEqual(t *testing.T, actual, expected Sample) (done bool) {
	if !expected.Equal(actual) {
		t.Errorf("Expected Sample '%s' but got '%s'", expected.String(), actual.String())
	}
	return true
}

func sortSamples(samples []Sample) {
	slices.SortFunc(
		samples,
		func(s1 Sample, s2 Sample) int {
			return strings.Compare(s1.String(), s2.String())
		},
	)
}

func testSamplesEqual(t *testing.T, actuals []Sample, expecteds []Sample) {
	testSamples(t, []AssertSamples{AssertSamplesEqual}, actuals, expecteds)
}

func testSamples(t *testing.T, asserters []AssertSamples, actuals []Sample, expecteds []Sample) {
	if len(actuals) != len(expecteds) {
		t.Errorf("Expected %d Samples, but got %d", len(expecteds), len(actuals))
	}

	sortSamples(actuals)
	sortSamples(expecteds)

	for i, actual := range actuals {
		expected := expecteds[i]

		for _, asserter := range asserters {
			if asserter(t, actual, expected) {
				break
			}
		}
	}
}
