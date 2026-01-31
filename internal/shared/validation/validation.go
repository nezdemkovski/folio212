package validation

import (
	"fmt"
	"strings"
)

func ValidateNonEmpty(fieldName, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}
