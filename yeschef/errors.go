package yeschef

import (
	"encoding/json"
	"fmt"
)

type UserVisibleError struct {
	Code    string                 // Error code for programmatic handling
	Message string                 // User-friendly error message
	Data    map[string]interface{} // Additional error context data
}

func (e *UserVisibleError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *UserVisibleError) JSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    e.Code,
			"message": e.Message,
			"data":    e.Data,
		},
	})
}

func NewUserVisibleError(code, message string, data map[string]interface{}) *UserVisibleError {
	return &UserVisibleError{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

func IsUserVisibleError(err error) bool {
	_, ok := err.(*UserVisibleError)
	return ok
}

func GetUserVisibleError(err error) *UserVisibleError {
	if uve, ok := err.(*UserVisibleError); ok {
		return uve
	}
	return nil
}
