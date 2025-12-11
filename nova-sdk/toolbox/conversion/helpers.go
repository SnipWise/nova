package conversion

import (
	"fmt"
	"strconv"
)

func StringToInt(str string) int {
	num, err := strconv.Atoi(str)
	if err != nil {
		fmt.Println("Cannot convert to int:", err)
		return 0
	}
	return num
}

func StringToFloat(str string) float64 {
	num, err := strconv.ParseFloat(str, 64)
	if err != nil {
		fmt.Println("Cannot convert to float:", err)
		return 0.0
	}
	return num
}

func StringToBool(str string) bool {
	val, err := strconv.ParseBool(str)
	if err != nil {
		fmt.Println("Cannot convert to bool:", err)
		return false
	}
	return val
}

// StringToIntErr converts a string to an integer and returns an error if conversion fails
func StringToIntErr(str string) (int, error) {
	num, err := strconv.Atoi(str)
	if err != nil {
		return 0, fmt.Errorf("cannot convert to int: %w", err)
	}
	return num, nil
}

// StringToFloatErr converts a string to a float64 and returns an error if conversion fails
func StringToFloatErr(str string) (float64, error) {
	num, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0.0, fmt.Errorf("cannot convert to float: %w", err)
	}
	return num, nil
}

// StringToBoolErr converts a string to a boolean and returns an error if conversion fails
func StringToBoolErr(str string) (bool, error) {
	val, err := strconv.ParseBool(str)
	if err != nil {
		return false, fmt.Errorf("cannot convert to bool: %w", err)
	}
	return val, nil
}

// StringToIntOrDefault converts a string to an integer, returning a custom default value if conversion fails
func StringToIntOrDefault(str string, defaultValue int) int {
	num, err := strconv.Atoi(str)
	if err != nil {
		return defaultValue
	}
	return num
}

// StringToFloatOrDefault converts a string to a float64, returning a custom default value if conversion fails
func StringToFloatOrDefault(str string, defaultValue float64) float64 {
	num, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return defaultValue
	}
	return num
}

// StringToBoolOrDefault converts a string to a boolean, returning a custom default value if conversion fails
func StringToBoolOrDefault(str string, defaultValue bool) bool {
	val, err := strconv.ParseBool(str)
	if err != nil {
		return defaultValue
	}
	return val
}

// IntToString converts an integer to a string
func IntToString(num int) string {
	return strconv.Itoa(num)
}

// FloatToString converts a float64 to a string
func FloatToString(num float64) string {
	return strconv.FormatFloat(num, 'f', -1, 64)
}

// BoolToString converts a boolean to a string
func BoolToString(val bool) string {
	return strconv.FormatBool(val)
}


