package noise

import (
	"fmt"
	"math"

	log "github.com/golang/glog"
)

type confInt struct{}

func (confInt) GaussianNoiseInterval(noisedValue float64, l0Sensitivity int64, lInfSensitivity, epsilon, confidenceLevel, delta float64) (float64, float64) {

	confidenceLevel = confidenceLevel / 100

	//delta := deltaForGaussian(sigma, l0Sensitivity, lInfSensitivity, epsilon)
	sigma := sigmaForGaussian(l0Sensitivity, lInfSensitivity, epsilon, delta)

	if err := checkArgsconfInt("GaussianNoiseInterval", l0Sensitivity, lInfSensitivity, epsilon, delta, confidenceLevel); err != nil {
		log.Fatalf("confInt.GaussianNoiseInterval( l0sensitivity %d, lInfSensitivity %f, epsilon %f, delta %e, confidenceLevel %f) checks failed with %v",
			l0Sensitivity, lInfSensitivity, epsilon, delta, confidenceLevel, err)
	}

	return getInterval(noisedValue, confidenceLevel, sigma)

}

func checkArgsconfInt(label string, l0Sensitivity int64, lInfSensitivity, epsilon, delta, confidenceLevel float64) error {

	if (confidenceLevel >= 0) || (confidenceLevel <= 1) {
		return fmt.Errorf("%s: confidenceLevel %f should be between 0 and 1", label, delta)
	}
	return checkArgsGaussian("GaussianNoiseInterval", l0Sensitivity, lInfSensitivity, epsilon, delta)
}

func getInterval(noisedValue float64, confidenceLevel float64, sigma float64) (float64, float64) {
	shiftingValue := sigma * math.Sqrt(2) * float64(math.Erfinv(2.00*confidenceLevel-1))

	X := noisedValue + shiftingValue
	Y := noisedValue - shiftingValue

	return X, Y

}
