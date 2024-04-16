package helpers

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// MSToTime converts a string of milliseconds to a ISO 8601-formatted string
func MSToTime(ms string) string {
	msInt, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse ms: %w", err).Error()
	}

	return time.Unix(0, msInt*int64(time.Millisecond)).Format("2006-01-02")
}

// GetRows converts a slice of House structs to a 2D slice of strings
func SliceToRow(h HouseSlice) [][]string {
	// Initialize empty list for rows
	var rows [][]string

	// Loop through each House
	for _, house := range h {
		// Get the house struct value as a reflect.Value
		v := reflect.ValueOf(house)
		// Check if it's a struct (should always be true here)
		if v.Kind() != reflect.Struct {
			fmt.Println("Unexpected type:", v.Kind())
			continue
		}

		// Empty slice to store house details
		houseRow := []string{}

		// Loop through each field of the struct
		for i := 0; i < v.NumField(); i++ {
			// Get the field value
			fieldValue := v.Field(i)
			// Convert the field value to a string (assuming all fields are strings)
			stringValue := fmt.Sprint(fieldValue.Interface())
			// Append the string value to the houseRow slice
			houseRow = append(houseRow, stringValue)
		}

		// Append the house details slice to rows
		rows = append(rows, houseRow)
	}

	return rows
}

// StructToSlice converts a struct to a slice of strings
func StructToSlice(s interface{}) []string {
	slice := []string{}
	fields := reflect.VisibleFields(reflect.TypeOf(s))
	for _, field := range fields {
		slice = append(slice, field.Name)
	}

	return slice
}

// JsonToMap converts a JSON-formatted string to a map[string]string
func JsonToMap(json string) map[string]string {
	jsonStr := strings.ReplaceAll(json, `\`, "")
	re := regexp.MustCompile(`"([^"]+)":"?([^"]+)"?`)
	matches := re.FindAllStringSubmatch(jsonStr, -1)
	dataMap := make(map[string]string)

	// Extract keys and values from matches
	for _, match := range matches {
		dataMap[match[1]] = match[2]
	}

	return dataMap
}
