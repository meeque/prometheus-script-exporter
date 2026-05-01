package main

import (
	"maps"
	"math"
	"reflect"
	"slices"
	"testing"
)

func (s1 Sample) Equal(s2 Sample) bool {
	return s1.Name == s2.Name &&
		maps.Equal(s1.Labels, s2.Labels) &&
		(s1.Value == s2.Value || (math.IsNaN(s1.Value) && math.IsNaN(s2.Value)))
}

type SampleAsserter interface {
	Assert(t *testing.T, sample *Sample) (match bool)
}

type ExactMatchAsserter struct {
	Sample *Sample
}

func (a ExactMatchAsserter) Assert(t *testing.T, sample *Sample) (match bool) {
	return a.Sample.Equal(*sample)
}

type MinDurationAsserter struct {
	Name   string
	Labels map[string]string
	Min    float64
	Max    float64
}

func (a MinDurationAsserter) Assert(t *testing.T, sample *Sample) (match bool) {
	match = a.Name == sample.Name &&
		reflect.DeepEqual(a.Labels, sample.Labels)
	if match {
		if sample.Value < a.Min || sample.Value > a.Max {
			t.Errorf("Expected sampled duration to be between %f and %f, but got %f", a.Min, a.Max, sample.Value)
		}
	}
	return
}

func assertSamples(t *testing.T, samples *[]Sample, expected *[]any) {
	asserters := []SampleAsserter{}

	for _, exp := range *expected {
		switch exp := exp.(type) {
		case SampleAsserter:
			asserters = append(asserters, exp)
		case Sample:
			asserters = append(asserters, ExactMatchAsserter{&exp})
		default:
			t.Logf("Unsupported type %T of expected Sample.", exp)
		}
	}

	assertSampleAsserters(t, samples, asserters)
}

func assertSampleAsserters(t *testing.T, samples *[]Sample, asserters []SampleAsserter) {
	if samples == nil && asserters == nil {
		return
	}
	if len(*samples) != len(asserters) {
		t.Errorf("Expected %d samples, got %d", len(asserters), len(*samples))
	}

	for _, asserter := range asserters {
		asserterMatchedASample := false
		for i, sample := range *samples {
			if asserter.Assert(t, &sample) {
				asserterMatchedASample = true
				*samples = slices.Delete(*samples, i, i+1)
				break
			}
		}
		if !asserterMatchedASample {
			t.Errorf("Asserter %s did not match any samples.", asserter)
		}
	}
	for _, sample := range *samples {
		t.Errorf("Unexpected sample %s was not matched by any asserter.", sample.String())
	}

}
