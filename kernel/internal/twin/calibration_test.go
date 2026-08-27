package twin

import (
	"errors"
	"math"
	"testing"
)

func TestCalibrateComputesBrierScoreAndRates(t *testing.T) {
	got, err := Calibrate([]Forecast{{ID: "a", Probability: 0.9, Outcome: true}, {ID: "b", Probability: 0.2, Outcome: false}})
	if err != nil {
		t.Fatal(err)
	}
	if got.Count != 2 || math.Abs(got.BrierScore-0.025) > 1e-12 || math.Abs(got.MeanProbability-0.55) > 1e-12 || got.ObservedRate != 0.5 {
		t.Fatalf("calibration=%+v", got)
	}
}

func TestCalibrateRejectsInvalidAndDuplicateFeedRecords(t *testing.T) {
	for name, feed := range map[string][]Forecast{
		"empty":        nil,
		"nan":          {{ID: "a", Probability: math.NaN()}},
		"negative":     {{ID: "a", Probability: -0.1}},
		"out of range": {{ID: "a", Probability: 1.1}},
		"duplicate":    {{ID: "a"}, {ID: "a"}},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Calibrate(feed); !errors.Is(err, ErrInvalidForecast) {
				t.Fatalf("error=%v", err)
			}
		})
	}
}
