//
// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package noise

import (
	"math"
	"testing"

	"github.com/grd/stat"
)

func TestGaussianStatistics(t *testing.T) {
	const numberOfSamples = 125000
	for _, tc := range []struct {
		l0Sensitivity                                   int64
		lInfSensitivity, epsilon, delta, mean, variance float64
	}{
		{
			l0Sensitivity:   1,
			lInfSensitivity: 1.0,
			epsilon:         ln3,
			delta:           1e-10,
			mean:            0.0,
			variance:        28.76478576660,
		},
		{
			l0Sensitivity:   1,
			lInfSensitivity: 1.0,
			epsilon:         ln3,
			delta:           1e-10,
			mean:            45941223.02107,
			variance:        28.76478576660,
		},
		{
			l0Sensitivity:   1,
			lInfSensitivity: 1.0,
			epsilon:         ln3,
			delta:           1e-10,
			mean:            0.0,
			variance:        28.76478576660,
		},
		{
			l0Sensitivity:   1,
			lInfSensitivity: 2.0,
			epsilon:         2.0 * ln3,
			delta:           1e-10,
			mean:            0.0,
			variance:        30.637955,
		},
		{
			l0Sensitivity:   2,
			lInfSensitivity: 1.0,
			epsilon:         2.0 * ln3,
			delta:           1e-10,
			mean:            0.0,
			variance:        15.318977,
		},
		{
			l0Sensitivity:   1,
			lInfSensitivity: 1.0,
			epsilon:         2 * ln3,
			delta:           1e-10,
			mean:            0.0,
			variance:        7.65948867798,
		},
		{
			l0Sensitivity:   1,
			lInfSensitivity: 1.0,
			epsilon:         ln3,
			delta:           1e-5,
			mean:            0.0,
			variance:        11.73597717285,
		},
	} {
		noisedSamples := make(stat.Float64Slice, numberOfSamples)
		for i := 0; i < numberOfSamples; i++ {
			noisedSamples[i] = gauss.AddNoiseFloat64(tc.mean, tc.l0Sensitivity, tc.lInfSensitivity, tc.epsilon, tc.delta)
		}
		sampleMean, sampleVariance := stat.Mean(noisedSamples), stat.Variance(noisedSamples)
		// Assuming that the Gaussian samples have a mean of 0 and the specified variance of tc.variance,
		// sampleMeanFloat64 and sampleMeanInt64 are approximately Gaussian distributed with a mean of 0
		// and standard deviation of sqrt(tc.variance⁻ / numberOfSamples).
		//
		// The meanErrorTolerance is set to the 99.9995% quantile of the anticipated distribution. Thus,
		// the test falsely rejects with a probability of 10⁻⁵.
		meanErrorTolerance := 4.41717 * math.Sqrt(tc.variance/float64(numberOfSamples))
		// Assuming that the Gaussian samples have the specified variance of tc.variance, sampleVarianceFloat64
		// and sampleVarianceInt64 are approximately Gaussian distributed with a mean of tc.variance and a
		// standard deviation of sqrt(2) * tc.variance / sqrt(numberOfSamples).
		//
		// The varianceErrorTolerance is set to the 99.9995% quantile of the anticipated distribution. Thus,
		// the test falsely rejects with a probability of 10⁻⁵.
		varianceErrorTolerance := 4.41717 * math.Sqrt2 * tc.variance / math.Sqrt(float64(numberOfSamples))

		if !nearEqual(sampleMean, tc.mean, meanErrorTolerance) {
			t.Errorf("float64 got mean = %f, want %f (parameters %+v)", sampleMean, tc.mean, tc)
		}
		if !nearEqual(sampleVariance, tc.variance, varianceErrorTolerance) {
			t.Errorf("float64 got variance = %f, want %f (parameters %+v)", sampleVariance, tc.variance, tc)
			sigma := sigmaForGaussian(tc.l0Sensitivity, tc.lInfSensitivity, tc.epsilon, tc.delta)
			t.Errorf("btw, true sigma is %f, squares to %f", sigma, sigma*sigma)
		}
	}
}

func TestSymmetricBinomialStatisitcs(t *testing.T) {
	const numberOfSamples = 125000
	for _, tc := range []struct {
		sqrtN  float64
		mean   float64
		stdDev float64
	}{
		{
			sqrtN:  1000.0,
			mean:   0.0,
			stdDev: 500.0,
		},
		{
			sqrtN:  1000000.0,
			mean:   0.0,
			stdDev: 500000.0,
		},
		{
			sqrtN:  1000000000.0,
			mean:   0.0,
			stdDev: 500000000.0,
		},
	} {
		binomialSamples := make(stat.IntSlice, numberOfSamples)
		for i := 0; i < numberOfSamples; i++ {
			binomialSamples[i] = symmetricBinomial(tc.sqrtN)
		}
		sampleMean, sampleVariance := stat.Mean(binomialSamples), stat.Variance(binomialSamples)
		// Assuming that the binomial samples have a mean of 0 and the specified standard deviation
		// of tc.stdDev, sampleMean is approximately Gaussian-distributed with a mean of 0
		// and standard deviation of tc.stdDev / sqrt(numberOfSamples).
		//
		// The meanErrorTolerance is set to the 99.9995% quantile of the anticipated distribution
		// of sampleMean. Thus, the test falsely rejects with a probability of 10⁻⁵.
		meanErrorTolerance := 4.41717 * tc.stdDev / math.Sqrt(float64(numberOfSamples))
		// Assuming that the binomial samples have the specified standard deviation of tc.stdDev,
		// sampleVariance is approximately Gaussian-distributed with a mean of tc.stdDev²
		// and a standard deviation of sqrt(2) * tc.stdDev² / sqrt(numberOfSamples).
		//
		// The varianceErrorTolerance is set to the 99.9995% quantile of the anticipated distribution
		// of sampleVariance. Thus, the test falsely rejects with a probability of 10⁻⁵.
		varianceErrorTolerance := 4.41717 * math.Sqrt2 * math.Pow(tc.stdDev, 2.0) / math.Sqrt(float64(numberOfSamples))

		if !nearEqual(sampleMean, tc.mean, meanErrorTolerance) {
			t.Errorf("got mean = %f, want %f (parameters %+v)", sampleMean, tc.mean, tc)
		}
		if !nearEqual(sampleVariance, math.Pow(tc.stdDev, 2.0), varianceErrorTolerance) {
			t.Errorf("got variance = %f, want %f (parameters %+v)", sampleVariance, math.Pow(tc.stdDev, 2.0), tc)
		}
	}
}

func TestDeltaForGaussian(t *testing.T) {
	for _, tc := range []struct {
		desc            string
		sigma           float64
		epsilon         float64
		l0Sensitivity   int64
		lInfSensitivity float64
		wantDelta       float64
		allowError      float64
	}{
		{
			desc:            "No noise added case",
			sigma:           0,
			epsilon:         1,
			l0Sensitivity:   1,
			lInfSensitivity: 1,
			// Attacker can deterministically verify all outputs of the Gaussian
			// mechanism when no noise is added.
			wantDelta: 1,
		},
		{
			desc:            "Overflow handling from large epsilon",
			sigma:           1,
			epsilon:         math.Inf(+1),
			l0Sensitivity:   1,
			lInfSensitivity: 1,
			// The full privacy leak is captured in the ε term.
			wantDelta: 0,
		},
		{
			desc:            "Overflow handling from large sensitivity",
			sigma:           1,
			epsilon:         1,
			l0Sensitivity:   1,
			lInfSensitivity: math.Inf(+1),
			// Infinite sensitivity cannot be hidden by finite noise.
			// No privacy guarantees.
			wantDelta: 1,
		},
		{
			desc:            "Underflow handling from low sensitivity",
			sigma:           1,
			epsilon:         1,
			l0Sensitivity:   1,
			lInfSensitivity: math.Nextafter(0, math.Inf(+1)),
			wantDelta:       0,
		},
		{
			desc:            "Correct value calculated",
			sigma:           10,
			epsilon:         0.1,
			l0Sensitivity:   1,
			lInfSensitivity: 1,
			wantDelta:       0.008751768145810,
			allowError:      1e-10,
		},
		{
			desc:            "Correct value calculated non-trivial lInfSensitivity",
			sigma:           20,
			l0Sensitivity:   1,
			lInfSensitivity: 2,
			epsilon:         0.1,
			wantDelta:       0.008751768145810,
			allowError:      1e-10,
		},
		{
			desc:            "Correct value calculated non-trivial l0Sensitivity",
			sigma:           20,
			l0Sensitivity:   4,
			lInfSensitivity: 1,
			epsilon:         0.1,
			wantDelta:       0.008751768145810,
			allowError:      1e-10,
		},
		{
			desc:            "Correct value calculated using typical epsilon",
			sigma:           10,
			l0Sensitivity:   1,
			lInfSensitivity: 5,
			epsilon:         math.Log(3),
			wantDelta:       0.004159742234000802,
			allowError:      1e-10,
		},
		{
			desc:            "Correct value calculated with epsilon = 0",
			sigma:           0.5,
			l0Sensitivity:   1,
			lInfSensitivity: 1,
			epsilon:         0,
			wantDelta:       0.6826894921370859,
			allowError:      1e-10,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			got := deltaForGaussian(tc.sigma, tc.l0Sensitivity, tc.lInfSensitivity, tc.epsilon)
			if math.Abs(got-tc.wantDelta) > tc.allowError {
				t.Errorf("Got delta: %1.11f, want delta: %1.11f", got, tc.wantDelta)
			}
		})
	}
}

func TestSigmaForGaussianInvertsDeltaForGaussian(t *testing.T) {
	// For these tests, we specify the value of sigma that we want to compute and
	// use DeltaForGaussian to determine the corresponding delta. We then verify
	// whether (given said delta) we can reconstruct sigma within the desired
	// tolerance. This validates that the function
	//   delta ↦ SigmaForGaussian(l2Sensitivity, epsilon, delta)
	// is an approximate inverse function of
	//   sigma ↦ DeltaForGuassian(sigma, l2Sensitivity, epsilon).

	for _, tc := range []struct {
		desc            string
		sigma           float64
		l0Sensitivity   int64
		lInfSensitivity float64
		epsilon         float64
	}{
		{
			desc:            "sigma smaller than l2Sensitivity",
			sigma:           0.3,
			l0Sensitivity:   1,
			lInfSensitivity: 0.5,
			epsilon:         0.5,
		},
		{
			desc:            "sigma larger than l2Sensitivity",
			sigma:           15,
			l0Sensitivity:   1,
			lInfSensitivity: 10,
			epsilon:         0.5,
		},
		{
			desc:            "sigma smaller non-trivial l0Sensitivity",
			sigma:           0.3,
			l0Sensitivity:   5,
			lInfSensitivity: 0.5,
			epsilon:         0.5,
		},
		{
			desc: "small delta",
			// Results in delta = 3.129776773173962e-141
			sigma:           500,
			l0Sensitivity:   1,
			lInfSensitivity: 10,
			epsilon:         0.5,
		},
		{
			desc:            "high lInfSensitivity",
			sigma:           1e102,
			l0Sensitivity:   1,
			lInfSensitivity: 1e100,
			epsilon:         0.1,
		},
		{
			desc:            "epsilon = 0",
			sigma:           0.5,
			l0Sensitivity:   1,
			lInfSensitivity: 1.0,
			epsilon:         0,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			deltaTight := deltaForGaussian(tc.sigma, tc.l0Sensitivity, tc.lInfSensitivity, tc.epsilon)
			gotSigma := sigmaForGaussian(tc.l0Sensitivity, tc.lInfSensitivity, tc.epsilon, deltaTight)
			if !(tc.sigma <= gotSigma && gotSigma <= (1+gaussianSigmaAccuracy)*tc.sigma) {
				t.Errorf("Got sigma: %f, want sigma in [%f, %f]", gotSigma, tc.sigma, (1+gaussianSigmaAccuracy)*tc.sigma)

			}
		})
	}
}

// This tests any logic that we need to special case for computing sigma (e.g.,
// precondition checking and boundary conditions).
func TestSigmaForGaussianWithDeltaOf1(t *testing.T) {
	got := sigmaForGaussian(1 /* l0 */, 1 /* lInf */, 0 /* ε */, 1 /* δ */)
	if got != 0 {
		t.Errorf("Got sigma: %f, want sigma: 0,", got)
	}
}

var thresholdGaussianTestCases = []struct {
	desc            string
	l0Sensitivity   int64
	lInfSensitivity float64
	epsilon         float64
	deltaNoise      float64
	deltaThreshold  float64
	threshold       float64
}{
	{
		desc:            "simple values",
		l0Sensitivity:   1,
		lInfSensitivity: 1,
		epsilon:         ln3,
		// deltaNoise is chosen to get a sigma of 1.
		deltaNoise: 0.10985556344445052,
		// 0.022750131948 is the 1-sided tail probability of landing more than 2
		// standard deviations from the mean of the Gaussian distribution.
		deltaThreshold: 0.022750131948,
		threshold:      3,
	},
	{
		desc:            "scale lInfSensitivity",
		l0Sensitivity:   1,
		lInfSensitivity: 0.5,
		epsilon:         ln3,
		// deltaNoise is chosen to get a sigma of 1.
		deltaNoise:     0.0041597422340007885,
		deltaThreshold: 0.000232629079,
		threshold:      4,
	},
	{
		desc:            "scale lInfSensitivity and sigma",
		l0Sensitivity:   1,
		lInfSensitivity: 2,
		epsilon:         ln3,
		// deltaNoise is chosen to get a sigma of 2.
		deltaNoise:     0.10985556344445052,
		deltaThreshold: 0.022750131948,
		threshold:      6,
	},
	{
		desc:            "scale l0Sensitivity",
		l0Sensitivity:   2,
		lInfSensitivity: 1,
		epsilon:         ln3,
		// deltaNoise is chosen to get a sigma of 1.
		deltaNoise:     0.26546844106038714,
		deltaThreshold: 0.022828893856,
		threshold:      3.275415487306,
	},
	{
		desc:            "small deltaThreshold",
		l0Sensitivity:   1,
		lInfSensitivity: 1,
		epsilon:         ln3,
		// deltaNoise is chosen to get a sigma of 1.
		deltaNoise: 0.10985556344445052,
		// 3e-5 is an approximate 1-sided tail probability of landing 4 standard
		// deviations from the mean of a Gaussian distribution.
		deltaThreshold: 3e-5,
		threshold:      5.012810811118,
	},
}

func TestThresholdGaussian(t *testing.T) {
	for _, tc := range thresholdGaussianTestCases {
		t.Run(tc.desc, func(t *testing.T) {
			gotThreshold := gauss.Threshold(tc.l0Sensitivity, tc.lInfSensitivity, tc.epsilon, tc.deltaNoise, tc.deltaThreshold)
			if math.Abs(gotThreshold-tc.threshold) > 1e-10 {
				t.Errorf("Got threshold: %0.12f, want threshold: %0.12f", gotThreshold, tc.threshold)
			}
		})
	}
}

func TestDeltaForThresholdGaussian(t *testing.T) {
	for _, tc := range thresholdGaussianTestCases {
		t.Run(tc.desc, func(t *testing.T) {
			gotDelta := gauss.(gaussian).DeltaForThreshold(tc.l0Sensitivity, tc.lInfSensitivity, tc.epsilon, tc.deltaNoise, tc.threshold)
			if math.Abs(gotDelta-tc.deltaThreshold) > 1e-10 {
				t.Errorf("Got delta: %0.12f, want delta: %0.12f", gotDelta, tc.deltaThreshold)
			}
		})
	}
}

func TestInverseCDFGaussian(t *testing.T) {
	for _, tc := range []struct {
		desc                         string
		sigma, confidenceLevel, want float64
	}{ //High precision tests
		{
			desc:            "High precision test, with random input",
			sigma:           1,
			confidenceLevel: 0.95,
			want:            1.64485362695,
		},
		{
			desc:            "High precision test, with random input",
			sigma:           2.342354,
			confidenceLevel: 0.8734521154362147425,
			want:            2.67698807013,
		},
		{
			desc:            "High precision test, with random input",
			sigma:           0.3,
			confidenceLevel: 0.75345892435835346586,
			want:            0.205624466704,
		},
		//Edge cases tests
		{
			desc:            "Edge case test with probability = 0",
			sigma:           0.3,
			confidenceLevel: 0,
			want:            math.Inf(-1),
		},
		{
			desc:            "Edge case test with probability = 1",
			sigma:           0.8,
			confidenceLevel: 1,
			want:            math.Inf(1),
		},
		{
			desc:            "Edge case test with low probability",
			sigma:           0.356,
			confidenceLevel: 0.05,
			want:            -0.585567891195,
		},
		{
			desc:            "Edge case test with high probability",
			sigma:           0.84,
			confidenceLevel: 0.99,
			want:            1.95413221419,
		},
		//Logical tests with probability of 0.5, it should return 0 = mean
		{
			desc:            "Logical test, with probability = 0.5",
			sigma:           0.3,
			confidenceLevel: 0.5,
			want:            0,
		},
		{
			desc:            "Logical test, with probability = 0.5",
			sigma:           0.8235243,
			confidenceLevel: 0.5,
			want:            0,
		},
	} {

		Zc := inverseCDFGaussian(tc.sigma, tc.confidenceLevel)
		if !(approxEqual(Zc, tc.want)) {
			t.Errorf(" TestInverseCDFGaussian(%f, %f) = %0.12f, want %0.12f, desc: %s", tc.sigma, tc.confidenceLevel, Zc, tc.want, tc.desc)

		}
	}
}

func TestConfidenceIntervalGaussian(t *testing.T) {
	// Tests for getConfidenceIntervalGaussian function
	for _, tc := range []struct {
		desc            string
		noisedValue     float64
		confidenceLevel float64
		sigma           float64
		want            ConfidenceIntervalFloat64
	}{
		// 4 random input tests.
		{
			desc:            "getConfidenceIntervalGaussian random input test",
			noisedValue:     21,
			sigma:           0.99999,
			confidenceLevel: 0.95,
			want:            ConfidenceIntervalFloat64{19.3551628216, 22.6448371784},
		},
		{
			desc:            "getConfidenceIntervalGaussian random input test",
			noisedValue:     40.003,
			sigma:           0.333,
			confidenceLevel: 0.888,
			want:            ConfidenceIntervalFloat64{39.5980851802, 40.4079148198},
		},
		{
			desc:            "getConfidenceIntervalGaussian random input test",
			noisedValue:     0.1,
			sigma:           9.123450004,
			confidenceLevel: 0.555,
			want:            ConfidenceIntervalFloat64{-1.16181152668, 1.36181152668},
		},
		{
			desc:            "getConfidenceIntervalGaussian random input test",
			noisedValue:     99.98989898,
			sigma:           15423235,
			confidenceLevel: 0.111,
			want:            ConfidenceIntervalFloat64{18835374.4248, -18835174.445},
		},
		// Confidence interval with confidence level of 0 and 1.

		{
			desc:            "Edge case test with probability = 0",
			noisedValue:     0,
			sigma:           1,
			confidenceLevel: 0, //For confidenceLevel = 0, -Infinity will be returned
			want:            ConfidenceIntervalFloat64{math.Inf(1), math.Inf(-1)},
		},
		{
			desc:            "Edge case test with probability = 0",
			noisedValue:     10,
			sigma:           100,
			confidenceLevel: 0, //For confidenceLevel = 0, -Infinity will be returned
			want:            ConfidenceIntervalFloat64{math.Inf(1), math.Inf(-1)},
		},
		{
			desc:            "Edge case test with probability = 1",
			noisedValue:     0,
			sigma:           1,
			confidenceLevel: 1, //For confidenceLevel = 1, Infinity will be returned
			want:            ConfidenceIntervalFloat64{math.Inf(-1), math.Inf(1)},
		},
		{
			desc:            "Edge case test with probability = 1",
			noisedValue:     100,
			sigma:           100,
			confidenceLevel: 1, //For confidenceLevel = 1, Infinity will be returned
			want:            ConfidenceIntervalFloat64{math.Inf(-1), math.Inf(1)},
		},
		//Near 0 and 1 confidence levels.
		{
			desc:            "Low confidence level",
			noisedValue:     100,
			sigma:           10,
			confidenceLevel: 0.001,
			want:            ConfidenceIntervalFloat64{130.902323062, 69.0976769383},
		},
		{
			desc:            "High confidence level",
			noisedValue:     100,
			sigma:           10,
			confidenceLevel: 0.9999,
			want:            ConfidenceIntervalFloat64{62.8098351454, 137.190164855},
		},
	} {
		result := getConfidenceIntervalGaussian(tc.noisedValue, tc.confidenceLevel, tc.sigma)
		if !approxEqual(result.LowerBound, tc.want.LowerBound) {
			t.Errorf("TestConfidenceIntervalGaussian(%f, %f, %f)=%0.10f, want %0.10f, desc %s, LowerBound is not equal",
				tc.noisedValue, tc.confidenceLevel, tc.sigma,
				result.LowerBound, tc.want.LowerBound, tc.desc)
		}
		if !approxEqual(result.UpperBound, tc.want.UpperBound) {
			t.Errorf("TestConfidenceIntervalLaplace(%f, %f, %f)=%0.10f, want %0.10f, desc %s, UpperBound is not equal",
				tc.noisedValue, tc.confidenceLevel, tc.sigma,
				result.UpperBound, tc.want.UpperBound, tc.desc)
		}
	}

}

func TestReturnConfidenceIntervalInt64(t *testing.T) {
	for _, tc := range []struct {
		desc                                        string
		noisedValue, l0Sensitivity, lInfSensitivity int64
		epsilon, delta, confidenceLevel             float64
		want                                        ConfidenceIntervalInt64
		wantErr                                     bool
	}{
		{
			desc:            "Random test",
			noisedValue:     70,
			l0Sensitivity:   6,
			lInfSensitivity: 10,
			epsilon:         0.3,
			delta:           0.1,
			confidenceLevel: 0.2,
			want:            ConfidenceIntervalInt64{110, 30}, //Values converted to int64
			wantErr:         false,
		},
		{
			desc:            "Random test",
			noisedValue:     1,
			l0Sensitivity:   1,
			lInfSensitivity: 15,
			epsilon:         0.5,
			delta:           0.9,
			confidenceLevel: 0.9,
			want:            ConfidenceIntervalInt64{-4, 6}, //Values converted to int64
			wantErr:         false,
		},
		//testing checkArgsConfidenceIntervalGaussian
		{
			desc:            "Testing confidence level bigger than 1",
			noisedValue:     1,
			l0Sensitivity:   1,
			lInfSensitivity: 15,
			epsilon:         0.5,
			delta:           0.9,
			confidenceLevel: 1.2,                            //The confidence level is bigger than 1, so it should return error
			want:            ConfidenceIntervalInt64{-4, 6}, //Random values, as test should return error because of the confidence level
			wantErr:         true,
		},
		{
			desc:            "Testing negative confidence level",
			noisedValue:     1,
			l0Sensitivity:   1,
			lInfSensitivity: 15,
			epsilon:         0.5,
			delta:           0.9,
			confidenceLevel: -5, //The confidence level is smaller than 0, so it should return error
			want:            ConfidenceIntervalInt64{-4, 6},
			wantErr:         true,
		},
		{
			desc:            "Testing negative l0Sensitivity",
			noisedValue:     1,
			l0Sensitivity:   -1, //The test should return an error if l0Sensitivity is not strictly positive
			lInfSensitivity: 15,
			epsilon:         0.5,
			delta:           0.9,
			confidenceLevel: 0.2,
			want:            ConfidenceIntervalInt64{-4, 6},
			wantErr:         true,
		},
		{
			desc:            "Testing zero l0Sensitivity",
			noisedValue:     1,
			l0Sensitivity:   0, //The test should return an error if l0Sensitivity is not strictly positive
			lInfSensitivity: 15,
			epsilon:         0.5,
			delta:           0.9,
			confidenceLevel: 0.2,
			want:            ConfidenceIntervalInt64{-4, 6},
			wantErr:         true,
		},
		{
			desc:            "Testing negative lInfSensitivity",
			noisedValue:     1,
			l0Sensitivity:   4,
			lInfSensitivity: -4, //lInfSensitivity should be strictly positive
			epsilon:         0.5,
			delta:           0.9,
			confidenceLevel: 0.2,
			want:            ConfidenceIntervalInt64{-4, 6},
			wantErr:         true,
		},
		{
			desc:            "Testing negative epsilon",
			noisedValue:     1,
			l0Sensitivity:   4,
			lInfSensitivity: 5,
			epsilon:         -0.05, //epsilon should be strictly positive
			delta:           0.9,
			confidenceLevel: 0.2,
			want:            ConfidenceIntervalInt64{-4, 6},
			wantErr:         true,
		},
		{
			desc:            "Testing negative dela",
			noisedValue:     1,
			l0Sensitivity:   4,
			lInfSensitivity: 5,
			epsilon:         0.05,
			delta:           -0.9, //delta should be strictly positive and smaller than 1
			confidenceLevel: 0.2,
			want:            ConfidenceIntervalInt64{-4, 6},
			wantErr:         true,
		},
		{
			desc:            "Testing bigger than 1 delta",
			noisedValue:     1,
			l0Sensitivity:   4,
			lInfSensitivity: 5,
			epsilon:         0.05,
			delta:           10, //delta should be strictly positive and smaller than 1
			confidenceLevel: 0.2,
			want:            ConfidenceIntervalInt64{-4, 6},
			wantErr:         true,
		},
	} {
		got, err := gauss.ReturnConfidenceIntervalInt64(tc.noisedValue, tc.l0Sensitivity, tc.lInfSensitivity,
			tc.epsilon, tc.delta, tc.confidenceLevel)
		if (err != nil) != tc.wantErr {
			t.Errorf("ReturnConfidenceIntervalInt64: when %s for err got %v", tc.desc, err)
			if got.LowerBound != tc.want.LowerBound {
				t.Errorf("TestReturnConfidenceIntervalInt64(%d, %d, %d, %f, %f, %f)=%d, want %d, desc %s, LowerBound is not equal",
					tc.noisedValue, tc.l0Sensitivity, tc.lInfSensitivity, tc.epsilon, tc.delta, tc.confidenceLevel,
					got.LowerBound, tc.want.LowerBound, tc.desc)
			}
			if got.UpperBound != tc.want.UpperBound {
				t.Errorf("TestReturnConfidenceIntervalInt64(%d, %d, %d, %f, %f, %f)=%d, want %d, desc %s, UpperBound is not equal",
					tc.noisedValue, tc.l0Sensitivity, tc.lInfSensitivity, tc.epsilon, tc.delta, tc.confidenceLevel,
					got.UpperBound, tc.want.UpperBound, tc.desc)
			}
		}
	}
}

func TestReturnConfidenceIntervalFloat64(t *testing.T) {
	for _, tc := range []struct {
		desc                                             string
		noisedValue                                      float64
		l0Sensitivity                                    int64
		lInfSensitivity, epsilon, delta, confidenceLevel float64
		want                                             ConfidenceIntervalFloat64
		wantErr                                          bool
	}{
		{
			desc:            "Random test",
			noisedValue:     70,
			l0Sensitivity:   5,
			lInfSensitivity: 36,
			epsilon:         0.8,
			delta:           0.8,
			confidenceLevel: 0.2,
			want:            ConfidenceIntervalFloat64{92.80911868743263, 47.19088131256736},
			wantErr:         false,
		},
		{
			desc:            "Random test",
			noisedValue:     60,
			l0Sensitivity:   1,
			lInfSensitivity: 5,
			epsilon:         0.333,
			delta:           0.9,
			confidenceLevel: 0.7,
			want:            ConfidenceIntervalFloat64{59.23887669725359, 60.76112330274641},
			wantErr:         false,
		},
		//testing checkArgsConfidenceIntervalGaussian
		{
			desc:            "Testing confidence level bigger than 1",
			noisedValue:     1,
			l0Sensitivity:   1,
			lInfSensitivity: 15,
			epsilon:         0.5,
			delta:           0.9,
			confidenceLevel: 1.2,                              //The confidence level is bigger than 1, so it should return error
			want:            ConfidenceIntervalFloat64{-4, 6}, //Random values, as test should return error because of the confidence level
			wantErr:         true,
		},
		{
			desc:            "Testing negative confidence level",
			noisedValue:     1,
			l0Sensitivity:   1,
			lInfSensitivity: 15,
			epsilon:         0.5,
			delta:           0.9,
			confidenceLevel: -5,                               //The confidence level is smaller than 0, so it should return error
			want:            ConfidenceIntervalFloat64{-4, 6}, //Random values
			wantErr:         true,
		},
		{
			desc:            "Testing negative l0Sensitivity",
			noisedValue:     1,
			l0Sensitivity:   -1, //The test should return an error if l0Sensitivity is not strictly positive
			lInfSensitivity: 15,
			epsilon:         0.5,
			delta:           0.9,
			confidenceLevel: 0.2,
			want:            ConfidenceIntervalFloat64{-4, 6}, //Random values
			wantErr:         true,
		},
		{
			desc:            "Testing zero l0Sensitivity",
			noisedValue:     1,
			l0Sensitivity:   0, //The test should return an error if l0Sensitivity is not strictly positive
			lInfSensitivity: 15,
			epsilon:         0.5,
			delta:           0.9,
			confidenceLevel: 0.2,
			want:            ConfidenceIntervalFloat64{-4, 6}, //Random values
			wantErr:         true,
		},
		{
			desc:            "Testing negative lInfSensitivity",
			noisedValue:     1,
			l0Sensitivity:   4,
			lInfSensitivity: -4, //lInfSensitivity should be strictly positive
			epsilon:         0.5,
			delta:           0.9,
			confidenceLevel: 0.2,
			want:            ConfidenceIntervalFloat64{-4, 6}, //Random values
			wantErr:         true,
		},
		{
			desc:            "Testing negative epsilon",
			noisedValue:     1,
			l0Sensitivity:   4,
			lInfSensitivity: 5,
			epsilon:         -0.05, //epsilon should be strictly positive
			delta:           0.9,
			confidenceLevel: 0.2,
			want:            ConfidenceIntervalFloat64{-4, 6}, //Random values
			wantErr:         true,
		},
		{
			desc:            "Testing negative dela",
			noisedValue:     1,
			l0Sensitivity:   4,
			lInfSensitivity: 5,
			epsilon:         0.05,
			delta:           -0.9, //delta should be strictly positive and smaller than 1
			confidenceLevel: 0.2,
			want:            ConfidenceIntervalFloat64{-4, 6}, //Random values
			wantErr:         true,
		},
		{
			desc:            "Testing bigger than 1 delta",
			noisedValue:     1,
			l0Sensitivity:   4,
			lInfSensitivity: 5,
			epsilon:         0.05,
			delta:           10, //delta should be strictly positive and smaller than 1
			confidenceLevel: 0.2,
			want:            ConfidenceIntervalFloat64{-4, 6}, //Random values
			wantErr:         true,
		},
		{
			desc:            "Testing infinite lInfSensitivity",
			noisedValue:     1,
			l0Sensitivity:   4,
			lInfSensitivity: math.Inf(1), //lInfSensitivity shouldn't be infinite
			epsilon:         0.05,
			delta:           0.3,
			confidenceLevel: 0.2,
			want:            ConfidenceIntervalFloat64{-4, 6}, //Random values
			wantErr:         true,
		}, {
			desc:            "Testing infinte epsilon",
			noisedValue:     1,
			l0Sensitivity:   4,
			lInfSensitivity: 5,
			epsilon:         math.Inf(1), //epsilon shouldn't be infinite
			delta:           0.4,
			confidenceLevel: 0.2,
			want:            ConfidenceIntervalFloat64{-4, 6}, //Random values
			wantErr:         true,
		},
	} {
		got, err := gauss.ReturnConfidenceIntervalFloat64(tc.noisedValue, tc.l0Sensitivity, tc.lInfSensitivity,
			tc.epsilon, tc.delta, tc.confidenceLevel)
		if (err != nil) != tc.wantErr {
			t.Errorf("ReturnConfidenceIntervalFloat64: when %s for err got %v", tc.desc, err)

			if !approxEqual(got.LowerBound, tc.want.LowerBound) {
				t.Errorf("TestReturnConfidenceIntervalFloat64(%f, %d, %f, %f, %f)=%0.10f, want %0.10f, desc %s, LowerBound is not equal",
					tc.noisedValue, tc.l0Sensitivity, tc.lInfSensitivity, tc.epsilon, tc.confidenceLevel,
					got.UpperBound, tc.want.UpperBound, tc.desc)
			}
			if !approxEqual(got.UpperBound, tc.want.UpperBound) {
				t.Errorf("TestReturnConfidenceIntervalFloat64(%f, %d, %f, %f, %f)=%0.10f, want %0.10f, desc %s, UpperBound is not equal",
					tc.noisedValue, tc.l0Sensitivity, tc.lInfSensitivity, tc.epsilon, tc.confidenceLevel,
					got.LowerBound, tc.want.LowerBound, tc.desc)
			}
		}
	}
}
