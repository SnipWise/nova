package tools

import (
	"errors"
	"reflect"
)

// extractToolCallback resolves a ToolCallback from a variadic callbacks slice at the given
// position. If not present or not assignable, the fallback is returned. Returns an error
// only when both the positional value and the fallback are nil.
func extractToolCallback(callbacks []any, position int, fallback ToolCallback) (ToolCallback, error) {
	if position < len(callbacks) && callbacks[position] != nil {
		if tc, ok := callbacks[position].(ToolCallback); ok {
			return tc, nil
		}
		// Try underlying function type (type aliases may not work in type assertions)
		v := reflect.ValueOf(callbacks[position])
		if v.Kind() == reflect.Func {
			t := v.Type()
			if t.NumIn() == 2 && t.NumOut() == 2 &&
				t.In(0).Kind() == reflect.String && t.In(1).Kind() == reflect.String &&
				t.Out(0).Kind() == reflect.String {
				if fn, ok := callbacks[position].(func(string, string) (string, error)); ok {
					return fn, nil
				}
			}
		}
	}
	if fallback != nil {
		return fallback, nil
	}
	return nil, errors.New("no tool callback provided: either pass ToolCallback parameter or set it via WithExecuteFn option")
}

// extractConfirmationCallback resolves a ConfirmationCallback from a variadic callbacks slice
// at the given position. Falls back to the provided fallback when not positionally supplied.
func extractConfirmationCallback(callbacks []any, position int, fallback ConfirmationCallback) (ConfirmationCallback, error) {
	if position < len(callbacks) && callbacks[position] != nil {
		if cc, ok := callbacks[position].(ConfirmationCallback); ok {
			return cc, nil
		}
		// Try underlying function type
		v := reflect.ValueOf(callbacks[position])
		if v.Kind() == reflect.Func {
			t := v.Type()
			if t.NumIn() == 2 && t.NumOut() == 1 &&
				t.In(0).Kind() == reflect.String && t.In(1).Kind() == reflect.String {
				if fn, ok := callbacks[position].(func(string, string) ConfirmationResponse); ok {
					return fn, nil
				}
			}
		}
	}
	if fallback != nil {
		return fallback, nil
	}
	return nil, errors.New("no confirmation callback provided: either pass ConfirmationCallback parameter or set it via WithConfirmationPromptFn option")
}
