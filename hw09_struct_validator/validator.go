package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("validation errors:\n")
	for i, err := range v {
		sb.WriteString(fmt.Sprintf("  %s: %v", err.Field, err.Err))
		if i < len(v)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

var (
	ErrNotStruct        = errors.New("value is not a struct")
	ErrInvalidValidator = errors.New("invalid validator format")
	ErrUnsupportedType  = errors.New("unsupported field type")
	ErrInvalidRegexp    = errors.New("invalid regexp pattern")
	ErrInvalidMinMax    = errors.New("minValue must be less than maxValue")
	ErrInvalidSliceType = errors.New("unsupported slice element type")
)

func Validate(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	var validationErrors ValidationErrors
	var err error
	validationErrors, err = validateStruct(val, "", validationErrors)
	if err != nil {
		return err
	}
	if len(validationErrors) == 0 {
		return nil
	}
	return validationErrors
}

func validateStruct(val reflect.Value, parentField string, errs ValidationErrors) (ValidationErrors, error) {
	typ := val.Type()
	var err error

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		if !field.IsExported() {
			continue
		}

		fieldName := field.Name
		if parentField != "" {
			fieldName = parentField + "." + fieldName
		}

		tag := field.Tag.Get("validate")
		if tag == "" {
			if fieldValue.Kind() == reflect.Struct && field.Type.Name() != "time.Time" {
				errs, err = validateStruct(fieldValue, fieldName, errs)
				if err != nil {
					return nil, err
				}
			}
			continue
		}

		if fieldValue.Kind() == reflect.Slice {
			errs, err = validateSlice(fieldValue, fieldName, tag, errs)
			if err != nil {
				return nil, err
			}
			continue
		}

		switch fieldValue.Kind() {
		case reflect.String:
			errs, err = validateString(fieldValue.String(), fieldName, tag, errs)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			errs, err = validateInt(fieldValue.Int(), fieldName, tag, errs)
		case reflect.Struct:
			if strings.Contains(tag, "nested") {
				errs, err = validateStruct(fieldValue, fieldName, errs)
			}
		case reflect.Invalid, reflect.Bool, reflect.Uint, reflect.Uint8, reflect.Uint16,
			reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64,
			reflect.Complex64, reflect.Complex128, reflect.Array, reflect.Chan, reflect.Func,
			reflect.Interface, reflect.Map, reflect.Pointer,
			reflect.Slice, reflect.UnsafePointer:
		default:
			return nil, ErrInvalidSliceType
		}
		if err != nil {
			return nil, err
		}
	}

	return errs, nil
}

func validateSlice(slice reflect.Value, fieldName, tag string, errs ValidationErrors) (ValidationErrors, error) {
	var err error
	if slice.Len() == 0 {
		return errs, nil
	}

	for i := 0; i < slice.Len(); i++ {
		elem := slice.Index(i)
		elemFieldName := fmt.Sprintf("%s[%d]", fieldName, i)

		switch elem.Kind() {
		case reflect.String:
			errs, err = validateString(elem.String(), elemFieldName, tag, errs)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			errs, err = validateInt(elem.Int(), elemFieldName, tag, errs)
		case reflect.Invalid, reflect.Bool, reflect.Uint, reflect.Uint8, reflect.Uint16,
			reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64,
			reflect.Complex64, reflect.Complex128, reflect.Array, reflect.Chan, reflect.Func,
			reflect.Interface, reflect.Map, reflect.Pointer,
			reflect.Slice, reflect.Struct, reflect.UnsafePointer:
		default:
			return nil, ErrInvalidSliceType
		}
		if err != nil {
			return nil, err
		}
	}

	return errs, nil
}

func validateString(value, fieldName, tag string, errs ValidationErrors) (ValidationErrors, error) {
	validators := strings.Split(tag, "|")

	for _, validator := range validators {
		validator = strings.TrimSpace(validator)
		if validator == "" {
			continue
		}

		parts := strings.SplitN(validator, ":", 2)
		if len(parts) != 2 {
			errs = append(errs, ValidationError{
				Field: fieldName,
				Err:   fmt.Errorf("%w: %s", ErrInvalidValidator, validator),
			})
			continue
		}

		rule := parts[0]
		ruleValue := parts[1]

		var validationError *ValidationError
		var err error
		switch rule {
		case "len":
			validationError, err = validateStringLen(value, ruleValue)
		case "regexp":
			validationError, err = validateStringRegexp(value, ruleValue)
		case "in":
			validationError = validateStringIn(value, ruleValue)
		case "minLen":
			validationError, err = validateStringMinLen(value, ruleValue)
		case "maxLen":
			validationError, err = validateStringMaxLen(value, ruleValue)
		default:
			err = fmt.Errorf("unknown string validator: %s", rule)
		}

		if err != nil {
			return nil, err
		}

		if validationError != nil {
			validationError.Field = fieldName
			errs = append(errs, *validationError)
		}
	}

	return errs, nil
}

func validateInt(value int64, fieldName, tag string, errs ValidationErrors) (ValidationErrors, error) {
	validators := strings.Split(tag, "|")

	for _, validator := range validators {
		validator = strings.TrimSpace(validator)
		if validator == "" {
			continue
		}

		parts := strings.SplitN(validator, ":", 2)
		if len(parts) != 2 {
			errs = append(errs, ValidationError{
				Field: fieldName,
				Err:   fmt.Errorf("%w: %s", ErrInvalidValidator, validator),
			})
			continue
		}

		rule := parts[0]
		ruleValue := parts[1]

		var validationError *ValidationError
		var err error
		switch rule {
		case "min":
			validationError, err = validateIntMin(value, ruleValue)
		case "max":
			validationError, err = validateIntMax(value, ruleValue)
		case "in":
			validationError, err = validateIntIn(value, ruleValue)
		case "range":
			validationError, err = validateIntRange(value, ruleValue)
		default:
			err = fmt.Errorf("unknown int validator: %s", rule)
		}

		if err != nil {
			return nil, err
		}

		if validationError != nil {
			validationError.Field = fieldName
			errs = append(errs, *validationError)
		}
	}

	return errs, nil
}

func validateStringLen(value, ruleValue string) (*ValidationError, error) {
	expectedLen, err := strconv.Atoi(ruleValue)
	if err != nil {
		return nil, ErrInvalidValidator
	}
	if len(value) != expectedLen {
		return &ValidationError{Err: fmt.Errorf("length must be %d, got %d", expectedLen, len(value))}, nil
	}
	return nil, nil
}

func validateStringRegexp(value, ruleValue string) (*ValidationError, error) {
	re, err := regexp.Compile(ruleValue)
	if err != nil {
		return nil, ErrInvalidRegexp
	}
	if !re.MatchString(value) {
		return &ValidationError{Err: fmt.Errorf("must match pattern %s", ruleValue)}, nil
	}
	return nil, nil
}

func validateStringIn(value, ruleValue string) *ValidationError {
	allowedValues := strings.Split(ruleValue, ",")
	for _, allowed := range allowedValues {
		if value == strings.TrimSpace(allowed) {
			return nil
		}
	}
	return &ValidationError{Err: fmt.Errorf("must be one of: %s", ruleValue)}
}

func validateStringMinLen(value, ruleValue string) (*ValidationError, error) {
	minLen, err := strconv.Atoi(ruleValue)
	if err != nil {
		return nil, ErrInvalidValidator
	}
	if len(value) < minLen {
		return &ValidationError{Err: fmt.Errorf("length must be at least %d, got %d", minLen, len(value))}, nil
	}
	return nil, nil
}

func validateStringMaxLen(value, ruleValue string) (*ValidationError, error) {
	maxLen, err := strconv.Atoi(ruleValue)
	if err != nil {
		return nil, ErrInvalidValidator
	}
	if len(value) > maxLen {
		return &ValidationError{Err: fmt.Errorf("length must be at most %d, got %d", maxLen, len(value))}, nil
	}
	return nil, nil
}

func validateIntMin(value int64, ruleValue string) (*ValidationError, error) {
	minValue, err := strconv.ParseInt(ruleValue, 10, 64)
	if err != nil {
		return nil, ErrInvalidValidator
	}
	if value < minValue {
		return &ValidationError{Err: fmt.Errorf("must be at least %d, got %d", minValue, value)}, nil
	}
	return nil, nil
}

func validateIntMax(value int64, ruleValue string) (*ValidationError, error) {
	maxValue, err := strconv.ParseInt(ruleValue, 10, 64)
	if err != nil {
		return nil, ErrInvalidValidator
	}
	if value > maxValue {
		return &ValidationError{Err: fmt.Errorf("must be at most %d, got %d", maxValue, value)}, nil
	}
	return nil, nil
}

func validateIntIn(value int64, ruleValue string) (*ValidationError, error) {
	allowedValues := strings.Split(ruleValue, ",")
	for _, allowed := range allowedValues {
		allowed = strings.TrimSpace(allowed)
		allowedInt, err := strconv.ParseInt(allowed, 10, 64)
		if err != nil {
			return nil, ErrInvalidValidator
		}
		if value == allowedInt {
			return nil, nil
		}
	}
	return &ValidationError{Err: fmt.Errorf("must be one of: %s", ruleValue)}, nil
}

func validateIntRange(value int64, ruleValue string) (*ValidationError, error) {
	parts := strings.Split(ruleValue, ",")
	if len(parts) != 2 {
		return nil, ErrInvalidValidator
	}

	minValue, err1 := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
	maxValue, err2 := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
	if err1 != nil || err2 != nil {
		return nil, ErrInvalidValidator
	}

	if minValue >= maxValue {
		return nil, ErrInvalidMinMax
	}

	if value < minValue || value > maxValue {
		return &ValidationError{
			Err: fmt.Errorf("must be between %d and %d, got %d", minValue, maxValue, value),
		}, nil
	}
	return nil, nil
}
