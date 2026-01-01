package validator

import (
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var (
	validate *validator.Validate
	once     sync.Once
)

func Init() {
	once.Do(func() {
		validate = validator.New()

		validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})

		validate.RegisterValidation("uuid", validateUUID)
		validate.RegisterValidation("notification_type", validateNotificationType)

		if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
			v.RegisterValidation("uuid", validateUUID)
			v.RegisterValidation("notification_type", validateNotificationType)
		}
	})
}

func Get() *validator.Validate {
	Init()
	return validate
}

func Validate(s interface{}) error {
	Init()
	return validate.Struct(s)
}

func validateUUID(fl validator.FieldLevel) bool {
	_, err := uuid.Parse(fl.Field().String())
	return err == nil
}

func validateNotificationType(fl validator.FieldLevel) bool {
	notifType := fl.Field().String()
	validTypes := []string{"info", "success", "warning", "error", "invitation"}
	for _, t := range validTypes {
		if notifType == t {
			return true
		}
	}
	return false
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func FormatValidationErrors(err error) []ValidationError {
	var errors []ValidationError

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, ValidationError{
				Field:   e.Field(),
				Message: getErrorMessage(e),
			})
		}
	}

	return errors
}

func getErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "min":
		return "Value is too short"
	case "max":
		return "Value is too long"
	case "uuid":
		return "Invalid UUID format"
	case "notification_type":
		return "Invalid type. Must be one of: info, success, warning, error, invitation"
	default:
		return "Invalid value"
	}
}

func ValidationMiddleware[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		var body T
		if err := c.ShouldBindJSON(&body); err != nil {
			if validationErrors, ok := err.(validator.ValidationErrors); ok {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":  "Validation failed",
					"errors": FormatValidationErrors(validationErrors),
				})
				c.Abort()
				return
			}

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request body",
			})
			c.Abort()
			return
		}

		c.Set("validated_body", body)
		c.Next()
	}
}

func GetValidatedBody[T any](c *gin.Context) (T, bool) {
	var zero T
	body, exists := c.Get("validated_body")
	if !exists {
		return zero, false
	}
	typed, ok := body.(T)
	return typed, ok
}
