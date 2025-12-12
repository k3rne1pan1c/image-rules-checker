package rules

import (
	"fmt"

	"github.com/k3rne1pan1c/image-checker/internal/model"
)

// Engine applies rules to image data
type Engine struct {
	config *Config
}

// NewEngine creates a new rule engine with the given configuration
func NewEngine(config *Config) *Engine {
	return &Engine{
		config: config,
	}
}

// CheckRules applies all enabled rules to an image and returns the results
func (e *Engine) CheckRules(img *model.Image) []model.RuleCheck {
	var checks []model.RuleCheck

	// Check minimum dimensions
	if e.config.Rules.MinDimensions != nil && e.config.Rules.MinDimensions.Enabled {
		check := e.checkMinDimensions(img)
		checks = append(checks, check)
	}

	// Check maximum dimensions
	if e.config.Rules.MaxDimensions != nil && e.config.Rules.MaxDimensions.Enabled {
		check := e.checkMaxDimensions(img)
		checks = append(checks, check)
	}

	// Check timestamp requirement
	if e.config.Rules.RequireTimestamp != nil && e.config.Rules.RequireTimestamp.Enabled {
		check := e.checkRequireTimestamp(img)
		checks = append(checks, check)
	}

	// Check GPS allowed/not allowed
	if e.config.Rules.GPSAllowed != nil && e.config.Rules.GPSAllowed.Enabled {
		check := e.checkGPSAllowed(img)
		checks = append(checks, check)
	}

	// Check orientation requirement
	if e.config.Rules.RequireOrientation != nil && e.config.Rules.RequireOrientation.Enabled {
		check := e.checkRequireOrientation(img)
		checks = append(checks, check)
	}

	return checks
}

func (e *Engine) checkMinDimensions(img *model.Image) model.RuleCheck {
	rule := e.config.Rules.MinDimensions
	if img.Image == nil {
		return model.RuleCheck{
			Name:    "min_dimensions",
			Passed:  false,
			Message: "image dimensions not available",
		}
	}

	width := img.Image.Dimensions.Width
	height := img.Image.Dimensions.Height

	if width < rule.Width || height < rule.Height {
		return model.RuleCheck{
			Name:    "min_dimensions",
			Passed:  false,
			Message: fmt.Sprintf("image too small: %dx%d (minimum: %dx%d)", width, height, rule.Width, rule.Height),
		}
	}

	return model.RuleCheck{
		Name:   "min_dimensions",
		Passed: true,
	}
}

func (e *Engine) checkMaxDimensions(img *model.Image) model.RuleCheck {
	rule := e.config.Rules.MaxDimensions
	if img.Image == nil {
		return model.RuleCheck{
			Name:    "max_dimensions",
			Passed:  false,
			Message: "image dimensions not available",
		}
	}

	width := img.Image.Dimensions.Width
	height := img.Image.Dimensions.Height

	if width > rule.Width || height > rule.Height {
		return model.RuleCheck{
			Name:    "max_dimensions",
			Passed:  false,
			Message: fmt.Sprintf("image too large: %dx%d (maximum: %dx%d)", width, height, rule.Width, rule.Height),
		}
	}

	return model.RuleCheck{
		Name:   "max_dimensions",
		Passed: true,
	}
}

func (e *Engine) checkRequireTimestamp(img *model.Image) model.RuleCheck {
	if img.EXIF == nil || img.EXIF.Timestamp == "" {
		return model.RuleCheck{
			Name:    "require_timestamp",
			Passed:  false,
			Message: "timestamp missing from EXIF data",
		}
	}

	return model.RuleCheck{
		Name:   "require_timestamp",
		Passed: true,
	}
}

func (e *Engine) checkGPSAllowed(img *model.Image) model.RuleCheck {
	rule := e.config.Rules.GPSAllowed
	hasGPS := img.EXIF != nil && img.EXIF.GPS != nil

	if rule.Allowed {
		// GPS is allowed - check if it's present (optional check)
		return model.RuleCheck{
			Name:   "gps_allowed",
			Passed: true,
		}
	} else {
		// GPS is not allowed - fail if present
		if hasGPS {
			return model.RuleCheck{
				Name:    "gps_allowed",
				Passed:  false,
				Message: "GPS not allowed but present",
			}
		}
		return model.RuleCheck{
			Name:   "gps_allowed",
			Passed: true,
		}
	}
}

func (e *Engine) checkRequireOrientation(img *model.Image) model.RuleCheck {
	if img.EXIF == nil || img.EXIF.Orientation == 0 {
		return model.RuleCheck{
			Name:    "require_orientation",
			Passed:  false,
			Message: "orientation missing from EXIF data",
		}
	}

	return model.RuleCheck{
		Name:   "require_orientation",
		Passed: true,
	}
}

