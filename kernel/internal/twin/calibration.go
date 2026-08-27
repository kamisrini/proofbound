package twin

import (
	"errors"
	"fmt"
	"math"
)

var ErrInvalidForecast = errors.New("twin: invalid forecast")

// Forecast is an observed forecast/outcome pair supplied by a calibration
// feed. Probability is the forecast chance of the boolean outcome being true.
type Forecast struct {
	ID          string
	Probability float64
	Outcome     bool
}

// Calibration is the aggregate Brier score and base rate for a forecast feed.
// Lower Brier scores are better; the score is in [0,1].
type Calibration struct {
	Count           int
	BrierScore      float64
	MeanProbability float64
	ObservedRate    float64
}

func Calibrate(forecasts []Forecast) (Calibration, error) {
	if len(forecasts) == 0 {
		return Calibration{}, fmt.Errorf("%w: empty feed", ErrInvalidForecast)
	}
	seen := make(map[string]struct{}, len(forecasts))
	var sumProbability, sumScore float64
	var observed int
	for i, forecast := range forecasts {
		if forecast.ID == "" || math.IsNaN(forecast.Probability) || math.IsInf(forecast.Probability, 0) || forecast.Probability < 0 || forecast.Probability > 1 {
			return Calibration{}, fmt.Errorf("%w: record %d", ErrInvalidForecast, i)
		}
		if _, ok := seen[forecast.ID]; ok {
			return Calibration{}, fmt.Errorf("%w: duplicate id %q", ErrInvalidForecast, forecast.ID)
		}
		seen[forecast.ID] = struct{}{}
		target := 0.0
		if forecast.Outcome {
			target = 1
			observed++
		}
		sumProbability += forecast.Probability
		delta := forecast.Probability - target
		sumScore += delta * delta
	}
	count := len(forecasts)
	return Calibration{Count: count, BrierScore: sumScore / float64(count), MeanProbability: sumProbability / float64(count), ObservedRate: float64(observed) / float64(count)}, nil
}
