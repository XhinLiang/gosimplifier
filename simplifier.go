package simplifier

import (
	"encoding/json"
	"reflect"
)

// Simplifier defines the interface for struct simplification.
type Simplifier interface {
	Simplify(original interface{}) (interface{}, error)
}

// simplifierImpl implements the Simplifier interface.
type simplifierImpl struct {
	rules map[string]interface{}
}

// NewSimplifier creates a new instance of simplifierImpl with the given rules based on the origin Simplifier.
// If the origin is nil, a new Simplifier is created without any base rules.
func NewSimplifier(origin Simplifier, rulesJSON string) (Simplifier, error) {
	rules := make(map[string]interface{})

	if origin != nil {
		// Copy the rules from the origin Simplifier
		rulesCopy, err := copyRules(origin)
		if err != nil {
			return nil, err
		}

		rules = rulesCopy
	}

	// Add or update the rules from the rulesJSON
	err := json.Unmarshal([]byte(rulesJSON), &rules)
	if err != nil {
		return nil, err
	}

	return &simplifierImpl{
		rules: rules,
	}, nil
}

// copyRules creates a deep copy of the rules from the origin Simplifier.
func copyRules(origin Simplifier) (map[string]interface{}, error) {
	originalRules := reflect.ValueOf(origin).Elem().FieldByName("rules")

	rulesBytes, err := json.Marshal(originalRules.Interface())
	if err != nil {
		return nil, err
	}

	rulesCopy := make(map[string]interface{})
	err = json.Unmarshal(rulesBytes, &rulesCopy)
	if err != nil {
		return nil, err
	}

	return rulesCopy, nil
}

// Simplify applies the rules to the original struct and returns a simplified copy.
func (s *simplifierImpl) Simplify(original interface{}) (interface{}, error) {
	copyValue := reflect.ValueOf(original)
	copyType := reflect.TypeOf(original)

	// Make a deep copy of the original value
	copy := reflect.New(copyType).Elem()
	copy = deepCopy(copy, copyValue)

	// Apply the rules recursively
	s.applyRules(copy, s.rules)

	return copy.Interface(), nil
}

// deepCopy recursively copies the values from src to dst.
func deepCopy(dst, src reflect.Value) reflect.Value {
	if src.Kind() == reflect.Ptr {
		if src.IsNil() {
			return dst
		}
		dst.Set(reflect.New(src.Type().Elem()))
		return deepCopy(dst.Elem(), src.Elem())
	}

	if src.Kind() == reflect.Struct {
		for i := 0; i < src.NumField(); i++ {
			srcField := src.Field(i)
			dstField := dst.Field(i)
			if srcField.CanInterface() {
				dstField.Set(deepCopy(dstField, srcField))
			}
		}
		return dst
	}

	if src.Kind() == reflect.Slice {
		dst.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Cap()))
		for i := 0; i < src.Len(); i++ {
			dst.Index(i).Set(deepCopy(dst.Index(i), src.Index(i)))
		}
		return dst
	}

	return src
}

// applyRules applies the given rules to the value recursively.
func (s *simplifierImpl) applyRules(value reflect.Value, rules map[string]interface{}) {
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return
		}
		value = value.Elem()
	}

	if value.Kind() == reflect.Struct {
		for fieldName, fieldValue := range rules {
			field := value.FieldByName(fieldName)
			if !field.IsValid() || !field.CanInterface() {
				continue
			}

			switch fieldValue := fieldValue.(type) {
			case []interface{}:
				if field.Kind() == reflect.Slice {
					for i := 0; i < field.Len(); i++ {
						item := field.Index(i)
						if item.Kind() == reflect.Ptr {
							item = item.Elem()
						}
						s.applyRules(item, fieldValue[i].(map[string]interface{}))
					}
				}
			case map[string]interface{}:
				s.applyRules(field, fieldValue)
			}
		}
	}
}
