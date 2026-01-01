package validator

import (
	"testing"
)

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		name string
		uuid string
		want bool
	}{
		{"valid uuid v4", "550e8400-e29b-41d4-a716-446655440000", true},
		{"valid uuid v1", "6ba7b810-9dad-11d1-80b4-00c04fd430c8", true},
		{"invalid format", "not-a-uuid", false},
		{"empty", "", false},
		{"partial uuid", "550e8400-e29b-41d4", false},
	}

	Init()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type testStruct struct {
				ID string `validate:"uuid"`
			}

			err := Validate(&testStruct{ID: tt.uuid})
			got := err == nil

			if got != tt.want {
				t.Errorf("validateUUID(%q) = %v, want %v", tt.uuid, got, tt.want)
			}
		})
	}
}

func TestValidateNotificationType(t *testing.T) {
	tests := []struct {
		name      string
		notifType string
		want      bool
	}{
		{"info", "info", true},
		{"success", "success", true},
		{"warning", "warning", true},
		{"error", "error", true},
		{"invitation", "invitation", true},
		{"invalid", "alert", false},
		{"empty", "", false},
		{"uppercase", "INFO", false},
	}

	Init()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type testStruct struct {
				Type string `validate:"notification_type"`
			}

			err := Validate(&testStruct{Type: tt.notifType})
			got := err == nil

			if got != tt.want {
				t.Errorf("validateNotificationType(%q) = %v, want %v", tt.notifType, got, tt.want)
			}
		})
	}
}

func TestFormatValidationErrors(t *testing.T) {
	Init()

	type testStruct struct {
		Title string `validate:"required" json:"title"`
		Type  string `validate:"required,notification_type" json:"type"`
	}

	err := Validate(&testStruct{Title: "", Type: ""})
	if err == nil {
		t.Fatal("expected validation error")
	}

	errors := FormatValidationErrors(err)
	if len(errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(errors))
	}
}

func TestGetErrorMessage(t *testing.T) {
	Init()

	type testStruct struct {
		Field string `validate:"required"`
	}

	err := Validate(&testStruct{Field: ""})
	if err == nil {
		t.Fatal("expected validation error")
	}

	errors := FormatValidationErrors(err)
	if len(errors) == 0 {
		t.Fatal("expected formatted errors")
	}

	if errors[0].Message != "This field is required" {
		t.Errorf("unexpected message: %s", errors[0].Message)
	}
}
