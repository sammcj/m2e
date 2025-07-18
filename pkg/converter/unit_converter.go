// Package converter provides unit conversion functionality
package converter

import (
	"fmt"
	"math"

	"github.com/martinlindhe/unit"
)

// UnitType represents different categories of units
type UnitType int

const (
	Length UnitType = iota
	Mass
	Volume
	Temperature
	Area
)

// UnitMatch represents a detected unit in text
type UnitMatch struct {
	Start      int
	End        int
	Value      float64
	Unit       string
	UnitType   UnitType
	Context    string
	Confidence float64
	IsCompound bool // true if this is a compound unit like "6-foot"
}

// ConversionResult represents the result of a unit conversion
type ConversionResult struct {
	MetricValue float64
	MetricUnit  string
	Formatted   string
	Confidence  float64
}

// ConversionPreferences holds user preferences for unit conversion
type ConversionPreferences struct {
	PreferWholeNumbers          bool
	MaxDecimalPlaces            int
	UseLocalizedUnits           bool
	TemperatureFormat           string  // "°C" or "degrees Celsius"
	UseSpaceBetweenValueAndUnit bool    // true: "5 kg", false: "5kg"
	RoundingThreshold           float64 // threshold for considering a value "close to whole" (default: 0.05)
}

// UnitConverter interface defines the contract for unit conversion
type UnitConverter interface {
	Convert(match UnitMatch) (ConversionResult, error)
	SetPrecision(unitType UnitType, precision int)
	SetPreferences(prefs ConversionPreferences)
}

// BasicUnitConverter implements the UnitConverter interface using martinlindhe/unit
type BasicUnitConverter struct {
	precision   map[UnitType]int
	preferences ConversionPreferences
}

// NewBasicUnitConverter creates a new BasicUnitConverter with default settings
func NewBasicUnitConverter() *BasicUnitConverter {
	return &BasicUnitConverter{
		precision: map[UnitType]int{
			Length:      1, // 1 decimal place for length
			Mass:        1, // 1 decimal place for mass
			Volume:      1, // 1 decimal place for volume
			Temperature: 0, // whole numbers for temperature
			Area:        1, // 1 decimal place for area
		},
		preferences: ConversionPreferences{
			PreferWholeNumbers:          true, // Changed back to true for better formatting
			MaxDecimalPlaces:            2,
			UseLocalizedUnits:           true,
			TemperatureFormat:           "°C",
			UseSpaceBetweenValueAndUnit: true,
			RoundingThreshold:           0.05, // More conservative threshold for whole number rounding
		},
	}
}

// SetPrecision sets the decimal precision for a specific unit type
func (c *BasicUnitConverter) SetPrecision(unitType UnitType, precision int) {
	c.precision[unitType] = precision
}

// SetPreferences sets the conversion preferences
func (c *BasicUnitConverter) SetPreferences(prefs ConversionPreferences) {
	c.preferences = prefs
}

// GetPreferences returns the current conversion preferences
func (c *BasicUnitConverter) GetPreferences() ConversionPreferences {
	return c.preferences
}

// Convert converts a unit match to metric equivalent
func (c *BasicUnitConverter) Convert(match UnitMatch) (ConversionResult, error) {
	switch match.UnitType {
	case Length:
		return c.convertLength(match)
	case Mass:
		return c.convertMass(match)
	case Volume:
		return c.convertVolume(match)
	case Temperature:
		return c.convertTemperature(match)
	case Area:
		return c.convertArea(match)
	default:
		return ConversionResult{}, fmt.Errorf("unsupported unit type: %v", match.UnitType)
	}
}

// convertLength converts imperial length units to metric
func (c *BasicUnitConverter) convertLength(match UnitMatch) (ConversionResult, error) {
	var metricValue float64
	var metricUnit string

	switch match.Unit {
	case "feet", "foot", "ft":
		metres := unit.Length(match.Value) * unit.Foot
		metricValue = metres.Meters()
		metricUnit = c.selectLengthUnit(metricValue, match.IsCompound, match.Unit)
	case "inches", "inch", "in":
		metres := unit.Length(match.Value) * unit.Inch
		metricValue = metres.Meters()
		metricUnit = c.selectLengthUnit(metricValue, match.IsCompound, match.Unit)
	case "yards", "yard", "yd":
		metres := unit.Length(match.Value) * unit.Yard
		metricValue = metres.Meters()
		metricUnit = c.selectLengthUnit(metricValue, match.IsCompound, match.Unit)
	case "miles", "mile", "mi":
		metres := unit.Length(match.Value) * unit.Mile
		metricValue = metres.Meters()
		metricUnit = c.selectLengthUnit(metricValue, match.IsCompound, match.Unit)
	default:
		return ConversionResult{}, fmt.Errorf("unsupported length unit: %s", match.Unit)
	}

	// Adjust value based on selected unit
	metricValue = c.adjustValueForUnit(metricValue, metricUnit)

	formatted := c.formatValue(metricValue, Length, metricUnit)

	return ConversionResult{
		MetricValue: metricValue,
		MetricUnit:  metricUnit,
		Formatted:   formatted,
		Confidence:  match.Confidence,
	}, nil
}

// convertMass converts imperial mass units to metric
func (c *BasicUnitConverter) convertMass(match UnitMatch) (ConversionResult, error) {
	var metricValue float64
	var metricUnit string

	switch match.Unit {
	case "pounds", "pound", "lbs", "lb":
		kg := unit.Mass(match.Value) * unit.AvoirdupoisPound
		metricValue = kg.Kilograms()
		metricUnit = c.selectMassUnit(metricValue)
	case "ounces", "ounce", "oz":
		kg := unit.Mass(match.Value) * unit.AvoirdupoisOunce
		metricValue = kg.Kilograms()
		metricUnit = c.selectMassUnit(metricValue)
	case "tons", "ton":
		// US short ton (short hundredweight)
		kg := unit.Mass(match.Value) * unit.ShortHundredweight * 20 // 20 short hundredweight = 1 short ton
		metricValue = kg.Kilograms()
		metricUnit = c.selectMassUnit(metricValue)
	default:
		return ConversionResult{}, fmt.Errorf("unsupported mass unit: %s", match.Unit)
	}

	// Adjust value based on selected unit
	metricValue = c.adjustValueForUnit(metricValue, metricUnit)

	formatted := c.formatValue(metricValue, Mass, metricUnit)

	return ConversionResult{
		MetricValue: metricValue,
		MetricUnit:  metricUnit,
		Formatted:   formatted,
		Confidence:  match.Confidence,
	}, nil
}

// convertVolume converts imperial volume units to metric
func (c *BasicUnitConverter) convertVolume(match UnitMatch) (ConversionResult, error) {
	var metricValue float64
	var metricUnit string

	switch match.Unit {
	case "gallons", "gallon", "gal":
		// US liquid gallon
		litres := unit.Volume(match.Value) * unit.USLiquidGallon
		metricValue = litres.Liters()
		metricUnit = c.selectVolumeUnit(metricValue)
	case "quarts", "quart", "qt":
		litres := unit.Volume(match.Value) * unit.USLiquidQuart
		metricValue = litres.Liters()
		metricUnit = c.selectVolumeUnit(metricValue)
	case "pints", "pint", "pt":
		litres := unit.Volume(match.Value) * unit.USLiquidPint
		metricValue = litres.Liters()
		metricUnit = c.selectVolumeUnit(metricValue)
	case "fluid ounces", "fluid ounce", "fl oz", "floz":
		litres := unit.Volume(match.Value) * unit.USFluidOunce
		metricValue = litres.Liters()
		metricUnit = c.selectVolumeUnit(metricValue)
	default:
		return ConversionResult{}, fmt.Errorf("unsupported volume unit: %s", match.Unit)
	}

	// Adjust value based on selected unit
	metricValue = c.adjustValueForUnit(metricValue, metricUnit)

	formatted := c.formatValue(metricValue, Volume, metricUnit)

	return ConversionResult{
		MetricValue: metricValue,
		MetricUnit:  metricUnit,
		Formatted:   formatted,
		Confidence:  match.Confidence,
	}, nil
}

// convertTemperature converts Fahrenheit to Celsius
func (c *BasicUnitConverter) convertTemperature(match UnitMatch) (ConversionResult, error) {
	switch match.Unit {
	case "fahrenheit", "°F", "F", "degrees fahrenheit":
		celsius := unit.FromFahrenheit(match.Value)
		metricValue := celsius.Celsius()

		formatted := c.formatValue(metricValue, Temperature, c.preferences.TemperatureFormat)

		return ConversionResult{
			MetricValue: metricValue,
			MetricUnit:  "°C",
			Formatted:   formatted,
			Confidence:  match.Confidence,
		}, nil
	default:
		return ConversionResult{}, fmt.Errorf("unsupported temperature unit: %s", match.Unit)
	}
}

// convertArea converts imperial area units to metric
func (c *BasicUnitConverter) convertArea(match UnitMatch) (ConversionResult, error) {
	var metricValue float64
	var metricUnit string

	switch match.Unit {
	case "square feet", "sq ft", "ft²":
		sqm := unit.Area(match.Value) * unit.SquareFoot
		metricValue = sqm.SquareMeters()
		metricUnit = c.selectAreaUnit(metricValue)
	case "acres", "acre":
		sqm := unit.Area(match.Value) * unit.Acre
		metricValue = sqm.SquareMeters()
		metricUnit = c.selectAreaUnit(metricValue)
	default:
		return ConversionResult{}, fmt.Errorf("unsupported area unit: %s", match.Unit)
	}

	// Adjust value based on selected unit
	metricValue = c.adjustValueForUnit(metricValue, metricUnit)

	formatted := c.formatValue(metricValue, Area, metricUnit)

	return ConversionResult{
		MetricValue: metricValue,
		MetricUnit:  metricUnit,
		Formatted:   formatted,
		Confidence:  match.Confidence,
	}, nil
}

// selectLengthUnit chooses the most appropriate metric length unit
func (c *BasicUnitConverter) selectLengthUnit(metres float64, isCompound bool, originalUnit string) string {
	absMetres := math.Abs(metres)

	// Special handling for inches - use mm for small values, cm for larger ones
	if originalUnit == "inches" || originalUnit == "inch" || originalUnit == "in" {
		if absMetres < 0.01 { // Less than 1 cm
			return "mm"
		} else if absMetres < 10.0 { // Less than 10 metres (about 400 inches)
			return "cm"
		}
	}

	if absMetres == 0.0 {
		if isCompound {
			return "metre" // Singular for compound units
		}
		return "metres" // Zero values use base unit
	} else if absMetres < 0.01 {
		return "mm"
	} else if absMetres < 1.0 {
		return "cm"
	} else if absMetres < 1000.0 {
		if isCompound {
			return "metre" // Singular for compound units like "2.7-metre"
		}
		return "metres"
	} else {
		return "km"
	}
}

// selectMassUnit chooses the most appropriate metric mass unit
func (c *BasicUnitConverter) selectMassUnit(kg float64) string {
	if kg < 0.001 {
		return "mg"
	} else if kg < 1.0 {
		return "g"
	} else if kg < 1000.0 {
		return "kg"
	} else {
		return "tonnes"
	}
}

// selectVolumeUnit chooses the most appropriate metric volume unit
func (c *BasicUnitConverter) selectVolumeUnit(litres float64) string {
	if litres < 0.001 {
		return "ml"
	} else if litres < 1.0 {
		return "ml"
	} else {
		return "litres"
	}
}

// selectAreaUnit chooses the most appropriate metric area unit
func (c *BasicUnitConverter) selectAreaUnit(sqm float64) string {
	if sqm < 10000.0 {
		return "m²"
	} else {
		return "hectares"
	}
}

// adjustValueForUnit adjusts the numeric value based on the selected unit
func (c *BasicUnitConverter) adjustValueForUnit(value float64, unit string) float64 {
	switch unit {
	case "mm":
		return value * 1000
	case "cm":
		return value * 100
	case "km":
		return value / 1000
	case "mg":
		return value * 1000000
	case "g":
		return value * 1000
	case "tonnes":
		return value / 1000
	case "ml":
		return value * 1000
	case "hectares":
		return value / 10000
	default:
		return value
	}
}

// formatValue formats the converted value according to preferences
func (c *BasicUnitConverter) formatValue(value float64, unitType UnitType, unit string) string {
	precision := c.precision[unitType]

	// Apply max decimal places limit
	if precision > c.preferences.MaxDecimalPlaces {
		precision = c.preferences.MaxDecimalPlaces
	}

	// Check if we should prefer whole numbers using configurable threshold
	if c.preferences.PreferWholeNumbers && math.Abs(value-math.Round(value)) < c.preferences.RoundingThreshold {
		return c.formatWithSpacing("%.0f", math.Round(value), unit)
	}

	// If not preferring whole numbers, but precision is 0, still format as whole number
	if precision == 0 {
		return c.formatWithSpacing("%.0f", math.Round(value), unit)
	}

	// For very small decimal parts, consider rounding to fewer decimal places
	if precision > 0 {
		// Check if we can use fewer decimal places without losing significant precision
		for p := 0; p < precision; p++ {
			roundedValue := math.Round(value*math.Pow(10, float64(p))) / math.Pow(10, float64(p))
			if math.Abs(value-roundedValue) < c.preferences.RoundingThreshold/10 {
				precision = p
				value = roundedValue
				break
			}
		}
	}

	// Format with specified precision
	format := fmt.Sprintf("%%.%df", precision)
	return c.formatWithSpacing(format, value, unit)
}

// formatWithSpacing applies spacing preferences between value and unit
func (c *BasicUnitConverter) formatWithSpacing(format string, value float64, unit string) string {
	formattedValue := fmt.Sprintf(format, value)

	// Special case for temperature units - no space before °C or °F for consistency with existing tests
	if unit == "°C" || unit == "°F" || unit == "degrees Celsius" {
		if unit == "degrees Celsius" {
			// For "degrees Celsius", we do want a space
			return fmt.Sprintf("%s %s", formattedValue, unit)
		}
		// For °C and °F, no space
		return fmt.Sprintf("%s%s", formattedValue, unit)
	}

	if c.preferences.UseSpaceBetweenValueAndUnit {
		return fmt.Sprintf("%s %s", formattedValue, unit)
	}
	return fmt.Sprintf("%s%s", formattedValue, unit)
}
