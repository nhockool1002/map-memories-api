package utils

import (
	"fmt"
	"reflect"
	"strings"

	"map-memories-api/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	
	// Register custom tag name function to use json tags for field names
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// ValidateStruct validates a struct and returns validation errors
func ValidateStruct(s interface{}) []models.ValidationError {
	var validationErrors []models.ValidationError
	
	err := validate.Struct(s)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, models.ValidationError{
				Field:   err.Field(),
				Message: getValidationMessage(err),
				Value:   err.Value(),
			})
		}
	}
	
	return validationErrors
}

// ValidateAndBindJSON validates and binds JSON request
func ValidateAndBindJSON(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return err
	}
	
	validationErrors := ValidateStruct(obj)
	if len(validationErrors) > 0 {
		c.JSON(400, models.ErrorResponseWithCode(
			"Validation failed",
			"VALIDATION_ERROR",
			validationErrors,
		))
		return fmt.Errorf("validation failed")
	}
	
	return nil
}

// ValidateAndBindQuery validates and binds query parameters
func ValidateAndBindQuery(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindQuery(obj); err != nil {
		return err
	}
	
	validationErrors := ValidateStruct(obj)
	if len(validationErrors) > 0 {
		c.JSON(400, models.ErrorResponseWithCode(
			"Validation failed",
			"VALIDATION_ERROR",
			validationErrors,
		))
		return fmt.Errorf("validation failed")
	}
	
	return nil
}

// getValidationMessage returns a human-readable validation error message
func getValidationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", fe.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", fe.Field(), fe.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", fe.Field(), fe.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", fe.Field(), fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", fe.Field(), fe.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", fe.Field(), fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", fe.Field(), fe.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", fe.Field(), fe.Param())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", fe.Field())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", fe.Field())
	case "numeric":
		return fmt.Sprintf("%s must be numeric", fe.Field())
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", fe.Field())
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", fe.Field())
	default:
		return fmt.Sprintf("%s is invalid", fe.Field())
	}
}

// IsValidUUID checks if a string is a valid UUID
func IsValidUUID(u string) bool {
	err := validate.Var(u, "uuid")
	return err == nil
}

// IsValidEmail checks if a string is a valid email
func IsValidEmail(email string) bool {
	err := validate.Var(email, "email")
	return err == nil
}

// IsValidLatitude checks if a float is a valid latitude
func IsValidLatitude(lat float64) bool {
	return lat >= -90 && lat <= 90
}

// IsValidLongitude checks if a float is a valid longitude
func IsValidLongitude(lng float64) bool {
	return lng >= -180 && lng <= 180
}