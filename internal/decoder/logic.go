package decoder

import "math"

const logicEpsilon = 1e-12

// LogicSteeringConfig defines CBS-style logic steering parameters.
type LogicSteeringConfig struct {
	Enabled            bool
	AttackTime         float64
	ReleaseTime        float64
	DominanceThreshold float64
	MaxBoost           float64
	MinGain            float64
}

// DefaultLogicSteeringConfig returns conservative logic steering defaults.
func DefaultLogicSteeringConfig() LogicSteeringConfig {
	return LogicSteeringConfig{
		Enabled:            false,
		AttackTime:         0.01,
		ReleaseTime:        0.2,
		DominanceThreshold: 0.55,
		MaxBoost:           1.6,
		MinGain:            0.4,
	}
}

func timeToCoeff(seconds float64, sampleRate int) float64 {
	if seconds <= 0 || sampleRate <= 0 {
		return 0
	}
	return math.Exp(-1.0 / (seconds * float64(sampleRate)))
}

func (d *SQDecoder) applyLogicSteering(lf, rf, lb, rb float64) (float64, float64, float64, float64) {
	energies := [4]float64{lf * lf, rf * rf, lb * lb, rb * rb}
	for i := 0; i < 4; i++ {
		env := d.logicEnv[i]
		energy := energies[i]
		if energy > env {
			d.logicEnv[i] = d.attackCoeff*env + (1.0-d.attackCoeff)*energy
		} else {
			d.logicEnv[i] = d.releaseCoeff*env + (1.0-d.releaseCoeff)*energy
		}
	}

	maxIdx := 0
	maxVal := d.logicEnv[0]
	sum := d.logicEnv[0] + d.logicEnv[1] + d.logicEnv[2] + d.logicEnv[3] + logicEpsilon
	for i := 1; i < 4; i++ {
		if d.logicEnv[i] > maxVal {
			maxVal = d.logicEnv[i]
			maxIdx = i
		}
	}

	dominance := maxVal / sum
	if dominance <= d.logicConfig.DominanceThreshold {
		return lf, rf, lb, rb
	}

	intensity := (dominance - d.logicConfig.DominanceThreshold) / (1.0 - d.logicConfig.DominanceThreshold)
	if intensity < 0 {
		intensity = 0
	} else if intensity > 1 {
		intensity = 1
	}

	boost := 1.0 + (d.logicConfig.MaxBoost-1.0)*intensity
	cut := 1.0 - (1.0-d.logicConfig.MinGain)*intensity

	gains := [4]float64{cut, cut, cut, cut}
	gains[maxIdx] = boost

	out := [4]float64{
		lf * gains[0],
		rf * gains[1],
		lb * gains[2],
		rb * gains[3],
	}

	preEnergy := energies[0] + energies[1] + energies[2] + energies[3]
	postEnergy := out[0]*out[0] + out[1]*out[1] + out[2]*out[2] + out[3]*out[3]
	if preEnergy > logicEpsilon && postEnergy > logicEpsilon {
		scale := math.Sqrt(preEnergy / postEnergy)
		out[0] *= scale
		out[1] *= scale
		out[2] *= scale
		out[3] *= scale
	}

	return out[0], out[1], out[2], out[3]
}
