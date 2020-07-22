package noise

import (
	"fmt"
	"math"

	log "github.com/golang/glog"
)

type confInt struct{}

func (confInt) GaussianNoiseInterval(noisedValue float64, l0Sensitivity int64, lInfSensitivity, epsilon, confidenceLevel, delta float64) (float64, float64) {
	//computing the st deviation with a given delta that we will check later
	sigma := sigmaForGaussian(l0Sensitivity, lInfSensitivity, epsilon, delta)
	//gettint the error in case the arguments are not valid
	if err := checkArgsconfInt("GaussianNoiseInterval", l0Sensitivity, lInfSensitivity, epsilon, delta, confidenceLevel); err != nil {
		log.Fatalf("confInt.GaussianNoiseInterval( l0sensitivity %d, lInfSensitivity %f, epsilon %f, delta %e, confidenceLevel %f) checks failed with %v",
			l0Sensitivity, lInfSensitivity, epsilon, delta, confidenceLevel, err)
	}

	return getInterval(noisedValue, confidenceLevel, sigma)

}

//checking the given arguments
func checkArgsconfInt(label string, l0Sensitivity int64, lInfSensitivity, epsilon, delta, confidenceLevel float64) error {
	//returning error in case the confidenceLevel is not between 0 and 1
	if (confidenceLevel >= 0) || (confidenceLevel <= 1) {
		return fmt.Errorf("%s: confidenceLevel %f should be between 0 and 1", label, delta)
	}
	return checkArgsGaussian("GaussianNoiseInterval", l0Sensitivity, lInfSensitivity, epsilon, delta)
}

//computing the confidence interval using the inverse error function
func getInterval(noisedValue float64, confidenceLevel float64, sigma float64) (float64, float64) {
	shiftingValue := sigma * math.Sqrt(2) * float64(math.Erfinv(2.00*confidenceLevel-1))

	lowerBound := noisedValue + shiftingValue
	upperBound := noisedValue - shiftingValue

	return lowerBound, upperBound

}
