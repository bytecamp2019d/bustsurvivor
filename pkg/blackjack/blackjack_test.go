package blackjack

import (
	"testing"
	"time"
)

func Test_BlackjackSurvival(t *testing.T) {
	jobStart := time.Now()
	numerator, denominator, _ := Survival(5, 21)
	t.Logf("Result is \033[1;92m%d/%d\033[0m, consumes %v", numerator, denominator, time.Since(jobStart))
	approximation := float64(numerator) / float64(denominator)
	t.Logf("Approximately equal to \033[1;92m%f\033[0m", approximation)
}
