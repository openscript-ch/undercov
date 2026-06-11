package coverage

import "fmt"

func CheckThreshold(path string, coverage, threshold float64) error {
	if coverage >= threshold {
		return nil
	}

	return fmt.Errorf("%s coverage %.2f%% is below threshold %.2f%%", path, coverage, threshold)
}

func CheckRegression(path string, previous, current float64) error {
	if current >= previous {
		return nil
	}

	return fmt.Errorf("%s coverage regressed from %.2f%% to %.2f%%", path, previous, current)
}
