package database

import "fmt"

// ValidateLatitude validates latitude bounds (-90 to 90)
func ValidateLatitude(lat float64) error {
	if lat < -90.0 || lat > 90.0 {
		return fmt.Errorf("latitude must be between -90 and 90 degrees, got: %f", lat)
	}
	return nil
}

// ValidateLongitude validates longitude bounds (-180 to 180)
func ValidateLongitude(lng float64) error {
	if lng < -180.0 || lng > 180.0 {
		return fmt.Errorf("longitude must be between -180 and 180 degrees, got: %f", lng)
	}
	return nil
}

// ValidateCoordinates validates both latitude and longitude
func ValidateCoordinates(lat, lng float64) error {
	if err := ValidateLatitude(lat); err != nil {
		return err
	}
	if err := ValidateLongitude(lng); err != nil {
		return err
	}
	return nil
}
