package types

import "fmt"

// TempRange represents temperature range in Celsius
type TempRange struct {
	MinC     *float64 `json:"min_c,omitempty"`
	MaxC     *float64 `json:"max_c,omitempty"`
	OptimalC *float64 `json:"optimal_c,omitempty"`
}

// IsValid validates temperature range
func (tr *TempRange) IsValid() error {
	if tr == nil {
		return nil
	}

	// Check absolute temperature limits (-273.1°C to 100°C)
	if tr.MinC != nil && (*tr.MinC < -273.1 || *tr.MinC > 100.0) {
		return fmt.Errorf("min temperature out of range: %.1f", *tr.MinC)
	}
	if tr.MaxC != nil && (*tr.MaxC < -273.1 || *tr.MaxC > 100.0) {
		return fmt.Errorf("max temperature out of range: %.1f", *tr.MaxC)
	}
	if tr.OptimalC != nil && (*tr.OptimalC < -273.1 || *tr.OptimalC > 100.0) {
		return fmt.Errorf("optimal temperature out of range: %.1f", *tr.OptimalC)
	}

	// Check logical order
	if tr.MinC != nil && tr.MaxC != nil && *tr.MinC > *tr.MaxC {
		return fmt.Errorf("min temperature (%.1f) cannot exceed max temperature (%.1f)", *tr.MinC, *tr.MaxC)
	}

	if tr.OptimalC != nil {
		if tr.MinC != nil && *tr.OptimalC < *tr.MinC {
			return fmt.Errorf("optimal temperature (%.1f) cannot be below min temperature (%.1f)", *tr.OptimalC, *tr.MinC)
		}
		if tr.MaxC != nil && *tr.OptimalC > *tr.MaxC {
			return fmt.Errorf("optimal temperature (%.1f) cannot exceed max temperature (%.1f)", *tr.OptimalC, *tr.MaxC)
		}
	}

	return nil
}

// PHRange represents pH range (0-14 scale)
type PHRange struct {
	MinPH     *float64 `json:"min_ph,omitempty"`
	MaxPH     *float64 `json:"max_ph,omitempty"`
	OptimalPH *float64 `json:"optimal_ph,omitempty"`
}

// IsValid validates pH range
func (pr *PHRange) IsValid() error {
	if pr == nil {
		return nil
	}

	// Check pH scale limits (0-14)
	if pr.MinPH != nil && (*pr.MinPH < 0 || *pr.MinPH > 14) {
		return fmt.Errorf("min pH out of range: %.1f", *pr.MinPH)
	}
	if pr.MaxPH != nil && (*pr.MaxPH < 0 || *pr.MaxPH > 14) {
		return fmt.Errorf("max pH out of range: %.1f", *pr.MaxPH)
	}
	if pr.OptimalPH != nil && (*pr.OptimalPH < 0 || *pr.OptimalPH > 14) {
		return fmt.Errorf("optimal pH out of range: %.1f", *pr.OptimalPH)
	}

	// Check logical order
	if pr.MinPH != nil && pr.MaxPH != nil && *pr.MinPH > *pr.MaxPH {
		return fmt.Errorf("min pH (%.1f) cannot exceed max pH (%.1f)", *pr.MinPH, *pr.MaxPH)
	}

	if pr.OptimalPH != nil {
		if pr.MinPH != nil && *pr.OptimalPH < *pr.MinPH {
			return fmt.Errorf("optimal pH (%.1f) cannot be below min pH (%.1f)", *pr.OptimalPH, *pr.MinPH)
		}
		if pr.MaxPH != nil && *pr.OptimalPH > *pr.MaxPH {
			return fmt.Errorf("optimal pH (%.1f) cannot exceed max pH (%.1f)", *pr.OptimalPH, *pr.MaxPH)
		}
	}

	return nil
}

// SizeRange represents physical size range in meters
type SizeRange struct {
	MinM     *float64 `json:"min_m,omitempty"`
	TypicalM *float64 `json:"typical_m,omitempty"`
	MaxM     *float64 `json:"max_m,omitempty"`
}

// IsValid validates size range
func (sr *SizeRange) IsValid() error {
	if sr == nil {
		return nil
	}

	// Check non-negative values
	if sr.MinM != nil && *sr.MinM < 0 {
		return fmt.Errorf("min size cannot be negative: %.2f", *sr.MinM)
	}
	if sr.TypicalM != nil && *sr.TypicalM < 0 {
		return fmt.Errorf("typical size cannot be negative: %.2f", *sr.TypicalM)
	}
	if sr.MaxM != nil && *sr.MaxM < 0 {
		return fmt.Errorf("max size cannot be negative: %.2f", *sr.MaxM)
	}

	// Check logical order
	if sr.MinM != nil && sr.TypicalM != nil && *sr.MinM > *sr.TypicalM {
		return fmt.Errorf("min size (%.2f) cannot exceed typical size (%.2f)", *sr.MinM, *sr.TypicalM)
	}
	if sr.TypicalM != nil && sr.MaxM != nil && *sr.TypicalM > *sr.MaxM {
		return fmt.Errorf("typical size (%.2f) cannot exceed max size (%.2f)", *sr.TypicalM, *sr.MaxM)
	}
	if sr.MinM != nil && sr.MaxM != nil && *sr.MinM > *sr.MaxM {
		return fmt.Errorf("min size (%.2f) cannot exceed max size (%.2f)", *sr.MinM, *sr.MaxM)
	}

	return nil
}

// Helper functions for creating pointers
func Float64Ptr(f float64) *float64 {
	return &f
}

func IntPtr(i int) *int {
	return &i
}

func StringPtr(s string) *string {
	return &s
}

func BoolPtr(b bool) *bool {
	return &b
}
