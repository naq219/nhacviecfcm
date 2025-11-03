package db

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/pocketbase/dbx"
)

// CustomMapper allows custom parsing for specific fields
type CustomMapper interface {
	MapFromDB(value string) (interface{}, error)
}

// MapperConfig holds optional configuration
type MapperConfig struct {
	AllowPartialMap bool                    // Allow missing fields without error
	CustomMappers   map[string]CustomMapper // field name -> custom mapper
	RequiredFields  []string                // Field names that must not be zero value
}

var (
	fieldCacheMap = make(map[string]map[string]*FieldInfo)
	cacheMutex    sync.RWMutex
)

type FieldInfo struct {
	Index int
	Type  reflect.Kind
	Kind  reflect.Type
	DBTag string
}

// MapNullStringMapToStruct maps dbx.NullStringMap to any struct T.
// Uses reflection to automatically parse field types based on db struct tags.
// Supports: bool, int*, uint*, float*, string, time.Time, structs (JSON), slices, maps.
func MapNullStringMapToStruct[T any](m dbx.NullStringMap) (*T, error) {
	return MapNullStringMapToStructWithConfig[T](m, &MapperConfig{})
}

// MapNullStringMapToStructWithConfig maps dbx.NullStringMap to struct T with optional configuration.
// Supports custom mappers for specific fields and required field validation.
// Example:
//
//	user, err := MapNullStringMapToStructWithConfig[User](raw, &MapperConfig{
//	    RequiredFields: []string{"ID", "Email"},
//	    CustomMappers: map[string]CustomMapper{"Status": statusMapper},
//	})
func MapNullStringMapToStructWithConfig[T any](m dbx.NullStringMap, cfg *MapperConfig) (*T, error) {
	if cfg == nil {
		cfg = &MapperConfig{}
	}

	var result T
	v := reflect.ValueOf(&result).Elem()
	fieldCache := getFieldCache[T]()

	for dbTag, fieldInfo := range fieldCache {
		if !m[dbTag].Valid {
			continue
		}

		value := m[dbTag].String
		fieldVal := v.Field(fieldInfo.Index)
		fieldName := v.Type().Field(fieldInfo.Index).Name

		// Check for custom mapper first
		if customMapper, ok := cfg.CustomMappers[fieldName]; ok {
			mapped, err := customMapper.MapFromDB(value)
			if err != nil {
				return nil, fmt.Errorf("custom mapper failed for field %s: %w", fieldName, err)
			}
			fieldVal.Set(reflect.ValueOf(mapped))
			continue
		}

		// Standard type mapping
		if err := mapFieldValue(fieldVal, fieldInfo.Kind, value, fieldName); err != nil {
			return nil, err
		}
	}

	// Validate required fields
	// Collects all missing/empty fields at once for better error reporting
	if err := validateRequiredFields(v, cfg.RequiredFields); err != nil {
		return nil, err
	}

	return &result, nil
}

// mapFieldValue maps a string value to a reflect.Value based on the target type.
// Handles primitive types (bool, int, uint, float, string), time.Time with multiple formats,
// and complex types (structs, slices, maps) using JSON unmarshaling.
// Returns error if parsing fails.
func mapFieldValue(fieldVal reflect.Value, fieldType reflect.Type, value string, fieldName string) error {
	switch fieldType.Kind() {
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid bool for field %s: %s", fieldName, value)
		}
		fieldVal.SetBool(b)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i64, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid int for field %s: %s", fieldName, value)
		}
		fieldVal.SetInt(i64)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u64, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid uint for field %s: %s", fieldName, value)
		}
		fieldVal.SetUint(u64)

	case reflect.Float32, reflect.Float64:
		f64, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float for field %s: %s", fieldName, value)
		}
		fieldVal.SetFloat(f64)

	case reflect.String:
		fieldVal.SetString(value)

	case reflect.Struct:
		if fieldType == reflect.TypeOf(time.Time{}) {
			t, err := parseTime(value)
			if err != nil {
				return fmt.Errorf("invalid time for field %s: %s", fieldName, value)
			}
			fieldVal.Set(reflect.ValueOf(t))
		} else {
			// Try JSON unmarshal for other structs
			if err := json.Unmarshal([]byte(value), fieldVal.Addr().Interface()); err != nil {
				log.Printf("Warning: failed to unmarshal JSON for field %s: %v", fieldName, err)
			}
		}

	case reflect.Slice, reflect.Map:
		// Try JSON unmarshal for complex types
		if err := json.Unmarshal([]byte(value), fieldVal.Addr().Interface()); err != nil {
			return fmt.Errorf("invalid JSON for field %s: %w", fieldName, err)
		}

	default:
		log.Printf("Warning: unsupported type for field %s: %v", fieldName, fieldType.Kind())
	}

	return nil
}

// parseTime attempts to parse a time string using multiple common formats.
// Supports: RFC3339Nano, RFC3339, ISO format, SQL datetime, PocketBase DateTime, and Go time.String() format.
// Returns error if none of the formats match.
func parseTime(value string) (time.Time, error) {
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05.999Z",    // PocketBase DateTime format with milliseconds
		"2006-01-02 15:04:05.000Z",    // PocketBase DateTime format
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05.9999999 -0700 MST", // Go time.String() format (PocketBase default)
	}

	for _, format := range formats {
		if t, err := time.Parse(format, value); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("time parsing failed for value: %s", value)
}

// getFieldCache builds and caches field metadata for struct T to optimize repeated mappings.
// Uses sync.RWMutex to ensure thread-safe access to the cache.
// Caches field index, type, and db tag for each field.
func getFieldCache[T any]() map[string]*FieldInfo {
	var t T
	typ := reflect.TypeOf(t)
	typeName := typ.String()

	// Check cache
	cacheMutex.RLock()
	if cached, ok := fieldCacheMap[typeName]; ok {
		cacheMutex.RUnlock()
		return cached
	}
	cacheMutex.RUnlock()

	// Build cache
	cache := make(map[string]*FieldInfo)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag == "" {
			continue
		}
		cache[dbTag] = &FieldInfo{
			Index: i,
			Type:  field.Type.Kind(),
			Kind:  field.Type,
			DBTag: dbTag,
		}
	}

	// Store in cache
	cacheMutex.Lock()
	fieldCacheMap[typeName] = cache
	cacheMutex.Unlock()

	return cache
}

// ClearFieldCache clears the field cache (useful for testing)
func ClearFieldCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	clear(fieldCacheMap)
}

// validateRequiredFields ensures all required fields are not empty or zero value.
// Collects all missing/empty fields and returns them in a single error message
// to help developers fix multiple issues at once.
func validateRequiredFields(v reflect.Value, required []string) error {
	if len(required) == 0 {
		return nil
	}

	missing := []string{}
	for _, fieldName := range required {
		f := v.FieldByName(fieldName)
		// Check if field exists and is not zero value
		if !f.IsValid() || f.IsZero() {
			missing = append(missing, fieldName)
		}
	}

	// If any required fields are missing/empty, return error with all of them
	if len(missing) > 0 {
		return fmt.Errorf("missing or empty required fields: %v", missing)
	}

	return nil
}
