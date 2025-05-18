package yeschef

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestUserVisibleErrorError(t *testing.T) {
	uve := &UserVisibleError{Code: "E1", Message: "msg"}
	if uve.Error() != "E1: msg" {
		t.Errorf("Error() = %s", uve.Error())
	}
}

func TestUserVisibleErrorJSON(t *testing.T) {
	uve := &UserVisibleError{Code: "E1", Message: "msg", Data: map[string]interface{}{"a": 1}}
	b, err := uve.JSON()
	if err != nil {
		t.Fatalf("JSON error: %v", err)
	}
	var parsed struct {
		Error struct {
			Code    string                 `json:"code"`
			Message string                 `json:"message"`
			Data    map[string]interface{} `json:"data"`
		} `json:"error"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if parsed.Error.Code != uve.Code || parsed.Error.Message != uve.Message || parsed.Error.Data["a"] != float64(1) {
		t.Errorf("unexpected JSON: %s", b)
	}
}

func TestNewUserVisibleError(t *testing.T) {
	uve := NewUserVisibleError("E1", "msg", map[string]interface{}{"a": 1})
	if uve.Code != "E1" || uve.Message != "msg" || uve.Data["a"] != 1 {
		t.Errorf("unexpected struct: %#v", uve)
	}
}

func TestIsAndGetUserVisibleError(t *testing.T) {
	err := errors.New("x")
	if IsUserVisibleError(err) {
		t.Error("expected false")
	}
	uve := NewUserVisibleError("E1", "msg", nil)
	if !IsUserVisibleError(uve) {
		t.Error("expected true")
	}
	if GetUserVisibleError(err) != nil {
		t.Error("expected nil")
	}
	if got := GetUserVisibleError(uve); got != uve {
		t.Errorf("got %#v", got)
	}
}
