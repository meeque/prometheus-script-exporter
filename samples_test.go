package main

import (
	"maps"
	"math"
	"reflect"
	"runtime"
	"slices"
	"testing"
)

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

func (s1 Sample) Equal(s2 Sample) bool {
	return s1.Name == s2.Name &&
		maps.Equal(s1.Labels, s2.Labels) &&
		(s1.Value == s2.Value || (math.IsNaN(s1.Value) && math.IsNaN(s2.Value)))
}

type MatchAndAssertSample func(sample *Sample) (match bool)

func NewMatchSample(expectedSample *Sample) MatchAndAssertSample {
	return func(sample *Sample) (match bool) {
		return expectedSample.Equal(*sample)
	}
}

func NewMatchSampleAndAssertValueRange(t *testing.T, name string, labels *map[string]string, min float64, max float64) MatchAndAssertSample {
	return func (sample *Sample) (match bool) {
		match = (sample.Name == name) &&
			reflect.DeepEqual(sample.Labels, *labels)
		if match {
			if sample.Value < min || sample.Value > max {
				t.Errorf("Expected sample value to be between %f and %f, but got %f", min, max, sample.Value)
			}
		}
		return
	}
}

func assertSamples(t *testing.T, samples *[]Sample, expected *[]any) {
	asserters := []MatchAndAssertSample{}

	for _, exp := range *expected {
		switch exp := exp.(type) {
		case MatchAndAssertSample:
			asserters = append(asserters, exp)
		case Sample:
			asserters = append(asserters, NewMatchSample(&exp))
		default:
			t.Logf("Unsupported type %T of expected Sample. Use either a Sample or a SampleAsserter!", exp)
		}
	}

	assertSampleAsserters(t, samples, &asserters)
}

func assertSampleAsserters(t *testing.T, samples *[]Sample, asserters *[]MatchAndAssertSample) {
	if samples == nil && asserters == nil {
		return
	}
	if len(*samples) != len(*asserters) {
		t.Errorf("Expected %d samples, got %d", len(*asserters), len(*samples))
	}

	for _, asserter := range *asserters {
		asserterMatchedASample := false
		for i, sample := range *samples {
			if asserter(&sample) {
				asserterMatchedASample = true
				*samples = slices.Delete(*samples, i, i+1)
				break
			}
		}
		if !asserterMatchedASample {
			t.Errorf("MatchAndAssertSample '%s' did not match any samples.", runtime.FuncForPC(reflect.ValueOf(asserter).Pointer()).Name())
		}
	}
	for _, sample := range *samples {
		t.Errorf("Unexpected sample %s was not matched by any asserter.", sample.String())
	}

}
