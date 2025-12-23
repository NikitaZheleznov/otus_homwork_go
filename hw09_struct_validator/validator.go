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
	validationErrors = validateStruct(val, "", validationErrors)

	if len(validationErrors) == 0 {
		return nil
	}
	return validationErrors
}

func validateStruct(val reflect.Value, parentField string, errs ValidationErrors) ValidationErrors {
	typ := val.Type()

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

		validateTag := field.Tag.Get("validate")
		if validateTag == "" {
			if fieldValue.Kind() == reflect.Struct && field.Type.Name() != "time.Time" {
				errs = validateStruct(fieldValue, fieldName, errs)
			}
			continue
		}

		if fieldValue.Kind() == reflect.Slice {
			errs = validateSlice(fieldValue, fieldName, validateTag, errs)
			continue
		}

		switch fieldValue.Kind() {
		case reflect.String:
			errs = validateString(fieldValue.String(), fieldName, validateTag, errs)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			errs = validateInt(fieldValue.Int(), fieldName, validateTag, errs)
		case reflect.Struct:
			if strings.Contains(validateTag, "nested") {
				errs = validateStruct(fieldValue, fieldName, errs)
			}
		case reflect.Invalid, reflect.Bool, reflect.Uint, reflect.Uint8, reflect.Uint16,
			reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64,
			reflect.Complex64, reflect.Complex128, reflect.Array, reflect.Chan, reflect.Func,
			reflect.Interface, reflect.Map, reflect.Pointer,
			reflect.Slice, reflect.UnsafePointer:
		default:
			errs = append(errs, ValidationError{
				Field: fieldName,
				Err:   fmt.Errorf("%w: %s", ErrUnsupportedType, fieldValue.Kind()),
			})
		}
	}

	return errs
}

func validateSlice(slice reflect.Value, fieldName, validateTag string, errs ValidationErrors) ValidationErrors {
	if slice.Len() == 0 {
		return errs
	}

	for i := 0; i < slice.Len(); i++ {
		elem := slice.Index(i)
		elemFieldName := fmt.Sprintf("%s[%d]", fieldName, i)

		switch elem.Kind() {
		case reflect.String:
			errs = validateString(elem.String(), elemFieldName, validateTag, errs)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			errs = validateInt(elem.Int(), elemFieldName, validateTag, errs)
		case reflect.Invalid, reflect.Bool, reflect.Uint, reflect.Uint8, reflect.Uint16,
			reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64,
			reflect.Complex64, reflect.Complex128, reflect.Array, reflect.Chan, reflect.Func,
			reflect.Interface, reflect.Map, reflect.Pointer,
			reflect.Slice, reflect.Struct, reflect.UnsafePointer:
		default:
			errs = append(errs, ValidationError{
				Field: elemFieldName,
				Err:   ErrInvalidSliceType,
			})
		}
	}

	return errs
}

func validateString(value, fieldName, validateTag string, errs ValidationErrors) ValidationErrors {
	validators := strings.Split(validateTag, "|")

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

		var err error
		switch rule {
		case "len":
			err = validateStringLen(value, ruleValue)
		case "regexp":
			err = validateStringRegexp(value, ruleValue)
		case "in":
			err = validateStringIn(value, ruleValue)
		case "minLen":
			err = validateStringMinLen(value, ruleValue)
		case "maxLen":
			err = validateStringMaxLen(value, ruleValue)
		default:
			err = fmt.Errorf("unknown string validator: %s", rule)
		}

		if err != nil {
			errs = append(errs, ValidationError{
				Field: fieldName,
				Err:   err,
			})
		}
	}

	return errs
}

func validateInt(value int64, fieldName, validateTag string, errs ValidationErrors) ValidationErrors {
	validators := strings.Split(validateTag, "|")

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

		var err error
		switch rule {
		case "min":
			err = validateIntMin(value, ruleValue)
		case "max":
			err = validateIntMax(value, ruleValue)
		case "in":
			err = validateIntIn(value, ruleValue)
		case "range":
			err = validateIntRange(value, ruleValue)
		default:
			err = fmt.Errorf("unknown int validator: %s", rule)
		}

		if err != nil {
			errs = append(errs, ValidationError{
				Field: fieldName,
				Err:   err,
			})
		}
	}

	return errs
}

func validateStringLen(value, ruleValue string) error {
	expectedLen, err := strconv.Atoi(ruleValue)
	if err != nil {
		return ErrInvalidValidator
	}
	if len(value) != expectedLen {
		return fmt.Errorf("length must be %d, got %d", expectedLen, len(value))
	}
	return nil
}

func validateStringRegexp(value, ruleValue string) error {
	re, err := regexp.Compile(ruleValue)
	if err != nil {
		return ErrInvalidRegexp
	}
	if !re.MatchString(value) {
		return fmt.Errorf("must match pattern %s", ruleValue)
	}
	return nil
}

func validateStringIn(value, ruleValue string) error {
	allowedValues := strings.Split(ruleValue, ",")
	for _, allowed := range allowedValues {
		if value == strings.TrimSpace(allowed) {
			return nil
		}
	}
	return fmt.Errorf("must be one of: %s", ruleValue)
}

func validateStringMinLen(value, ruleValue string) error {
	minLen, err := strconv.Atoi(ruleValue)
	if err != nil {
		return ErrInvalidValidator
	}
	if len(value) < minLen {
		return fmt.Errorf("length must be at least %d, got %d", minLen, len(value))
	}
	return nil
}

func validateStringMaxLen(value, ruleValue string) error {
	maxLen, err := strconv.Atoi(ruleValue)
	if err != nil {
		return ErrInvalidValidator
	}
	if len(value) > maxLen {
		return fmt.Errorf("length must be at most %d, got %d", maxLen, len(value))
	}
	return nil
}

func validateIntMin(value int64, ruleValue string) error {
	minValue, err := strconv.ParseInt(ruleValue, 10, 64)
	if err != nil {
		return ErrInvalidValidator
	}
	if value < minValue {
		return fmt.Errorf("must be at least %d, got %d", minValue, value)
	}
	return nil
}

func validateIntMax(value int64, ruleValue string) error {
	maxValue, err := strconv.ParseInt(ruleValue, 10, 64)
	if err != nil {
		return ErrInvalidValidator
	}
	if value > maxValue {
		return fmt.Errorf("must be at most %d, got %d", maxValue, value)
	}
	return nil
}

func validateIntIn(value int64, ruleValue string) error {
	allowedValues := strings.Split(ruleValue, ",")
	for _, allowed := range allowedValues {
		allowed = strings.TrimSpace(allowed)
		allowedInt, err := strconv.ParseInt(allowed, 10, 64)
		if err != nil {
			return ErrInvalidValidator
		}
		if value == allowedInt {
			return nil
		}
	}
	return fmt.Errorf("must be one of: %s", ruleValue)
}

func validateIntRange(value int64, ruleValue string) error {
	parts := strings.Split(ruleValue, ",")
	if len(parts) != 2 {
		return ErrInvalidValidator
	}

	minValue, err1 := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
	maxValue, err2 := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
	if err1 != nil || err2 != nil {
		return ErrInvalidValidator
	}

	if minValue >= maxValue {
		return ErrInvalidMinMax
	}

	if value < minValue || value > maxValue {
		return fmt.Errorf("must be between %d and %d, got %d", minValue, maxValue, value)
	}
	return nil
}
