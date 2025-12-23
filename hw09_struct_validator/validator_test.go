package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}

	testCase struct {
		name        string
		in          interface{}
		expectedErr error
		hasErrors   bool
		errCount    int
	}
)

func TestValidate(t *testing.T) {
	tests := getTestCases()
	for _, tt := range tests {
		t.Run(fmt.Sprintf("Case %s", tt.name), func(t *testing.T) {
			testCase := tt
			t.Parallel()
			err := Validate(testCase.in)
			runTestCase(t, testCase, err)
		})
	}
}

func getTestCases() []testCase {
	return []testCase{
		{
			name: "Valid User struct",
			in: User{
				ID:     "12345678-1234-1234-1234-123456789012",
				Name:   "John Doe",
				Age:    25,
				Email:  "test@example.com",
				Role:   "admin",
				Phones: []string{"12345678901", "10987654321"},
			},
			expectedErr: nil,
			hasErrors:   false,
		},
		{
			name: "Valid App struct",
			in: App{
				Version: "1.0.0",
			},
			expectedErr: nil,
			hasErrors:   false,
		},
		{
			name: "Valid Response struct",
			in: Response{
				Code: 200,
				Body: "OK",
			},
			expectedErr: nil,
			hasErrors:   false,
		},
		{
			name: "Token struct without validation tags",
			in: Token{
				Header:    []byte("header"),
				Payload:   []byte("payload"),
				Signature: []byte("signature"),
			},
			expectedErr: nil,
			hasErrors:   false,
		},
		{
			name: "Pointer to valid User",
			in: &User{
				ID:     "12345678-1234-1234-1234-123456789012",
				Name:   "John Doe",
				Age:    25,
				Email:  "test@example.com",
				Role:   "stuff",
				Phones: []string{"12345678901"},
			},
			expectedErr: nil,
			hasErrors:   false,
		},
		{
			name: "App with invalid version length",
			in: App{
				Version: "1.0",
			},
			hasErrors: true,
			errCount:  1,
		},
		{
			name: "Response with invalid code",
			in: Response{
				Code: 400,
				Body: "Bad Request",
			},
			hasErrors: true,
			errCount:  1,
		},
		{
			name: "User with invalid ID length",
			in: User{
				ID:     "invalid",
				Name:   "John Doe",
				Age:    25,
				Email:  "test@example.com",
				Role:   "admin",
				Phones: []string{"12345678901"},
			},
			hasErrors: true,
			errCount:  1,
		},
		{
			name: "User with invalid age (too young)",
			in: User{
				ID:     "12345678-1234-1234-1234-123456789012",
				Name:   "John Doe",
				Age:    16,
				Email:  "test@example.com",
				Role:   "admin",
				Phones: []string{"12345678901"},
			},
			hasErrors: true,
			errCount:  1,
		},
		{
			name: "User with invalid age (too old)",
			in: User{
				ID:     "12345678-1234-1234-1234-123456789012",
				Name:   "John Doe",
				Age:    51,
				Email:  "test@example.com",
				Role:   "admin",
				Phones: []string{"12345678901"},
			},
			hasErrors: true,
			errCount:  1,
		},
		{
			name: "User with invalid email",
			in: User{
				ID:     "12345678-1234-1234-1234-123456789012",
				Name:   "John Doe",
				Age:    25,
				Email:  "invalid",
				Role:   "admin",
				Phones: []string{"12345678901"},
			},
			hasErrors: true,
			errCount:  1,
		},
		{
			name: "User with invalid role",
			in: User{
				ID:     "12345678-1234-1234-1234-123456789012",
				Name:   "John Doe",
				Age:    25,
				Email:  "test@example.com",
				Role:   "invalid",
				Phones: []string{"12345678901"},
			},
			hasErrors: true,
			errCount:  1,
		},
		{
			name: "User with invalid phone length",
			in: User{
				ID:     "12345678-1234-1234-1234-123456789012",
				Name:   "John Doe",
				Age:    25,
				Email:  "test@example.com",
				Role:   "admin",
				Phones: []string{"123"},
			},
			hasErrors: true,
			errCount:  1,
		},
		{
			name: "User with multiple errors",
			in: User{
				ID:     "invalid",
				Name:   "John Doe",
				Age:    15,
				Email:  "invalid",
				Role:   "invalid",
				Phones: []string{"123"},
			},
			hasErrors: true,
			errCount:  5,
		},
		{
			name: "Empty slice in phones",
			in: User{
				ID:     "12345678-1234-1234-1234-123456789012",
				Name:   "John Doe",
				Age:    25,
				Email:  "test@example.com",
				Role:   "admin",
				Phones: []string{},
			},
			expectedErr: nil,
			hasErrors:   false,
		},
		{
			name: "Nil slice in phones",
			in: User{
				ID:     "12345678-1234-1234-1234-123456789012",
				Name:   "John Doe",
				Age:    25,
				Email:  "test@example.com",
				Role:   "admin",
				Phones: nil,
			},
			expectedErr: nil,
			hasErrors:   false,
		},
	}
}

func runTestCase(t *testing.T, tc testCase, err error) {
	t.Helper()
	if tc.expectedErr != nil {
		if !errors.Is(err, tc.expectedErr) {
			t.Errorf("expected error %v, got %v", tc.expectedErr, err)
		}
		return
	}
	if tc.hasErrors {
		checkHasErrors(t, tc, err)
		return
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func checkHasErrors(t *testing.T, tc testCase, err error) {
	t.Helper()
	if err == nil {
		t.Error("expected validation errors, got nil")
		return
	}
	var valErrs ValidationErrors
	if !errors.As(err, &valErrs) {
		t.Errorf("expected ValidationErrors, got %T: %v", err, err)
		return
	}
	if tc.errCount > 0 && len(valErrs) != tc.errCount {
		t.Errorf("expected %d validation errors, got %d", tc.errCount, len(valErrs))
		for i, ve := range valErrs {
			t.Errorf("  error %d: %s - %v", i, ve.Field, ve.Err)
		}
	}
}
