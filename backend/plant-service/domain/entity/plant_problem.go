package entity

import (
	"fmt"
	"time"
)

// PlantProblem represents a pest, disease, or deficiency affecting a plant
type PlantProblem struct {
	ProblemID   string    `json:"problem_id"`
	PlantID     string    `json:"plant_id"`
	ProblemType string    `json:"problem_type"` // pest, disease, deficiency, toxicity
	Severity    string    `json:"severity"`     // low, medium, high
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Validate validates the plant problem entity
func (pp *PlantProblem) Validate() error {
	if pp.PlantID == "" {
		return fmt.Errorf("plant_id is required")
	}

	if pp.ProblemType == "" {
		return fmt.Errorf("problem_type is required")
	}

	// Validate problem type
	validTypes := map[string]bool{
		"pest":        true,
		"disease":     true,
		"deficiency":  true,
		"toxicity":    true,
	}

	if !validTypes[pp.ProblemType] {
		return fmt.Errorf("invalid problem_type: %s (must be pest, disease, deficiency, or toxicity)", pp.ProblemType)
	}

	// Validate severity
	validSeverities := map[string]bool{
		"low":    true,
		"medium": true,
		"high":   true,
	}

	if !validSeverities[pp.Severity] {
		return fmt.Errorf("invalid severity: %s (must be low, medium, or high)", pp.Severity)
	}

	return nil
}
