package gosimplifier

import (
	"encoding/json"
	"reflect"
)

// Rule defines the rule structure for property removal and nested property rules.
type Rule struct {
	RemoveProperties    []string         `json:"remove_properties"`
	PropertySimplifiers map[string]*Rule `json:"property_simplifiers"`
}

// Simplifier defines the interface for struct simplification.
type Simplifier interface {
	Simplify(original interface{}) (interface{}, error)
}

// simplifierImpl implements the Simplifier interface.
type simplifierImpl struct {
	propertySimplifiers map[string]Ruler
}

type Ruler interface {
	applyRules(name interface{}, value reflect.Value)
}

type clearRuler struct {
}

var clearRulerIns = &clearRuler{}

// NewSimplifier creates a new instance of simplifierImpl with the given rules based on the origin Simplifier.
// If the origin is nil, a new Simplifier is created without any base rules.
func NewSimplifier(rulesJSON string) (Simplifier, error) {
	rule := &Rule{}
	err := json.Unmarshal([]byte(rulesJSON), rule)
	if err != nil {
		return nil, err
	}
	return newSimplifier0(rule)
}

func newSimplifier0(rule *Rule) (*simplifierImpl, error) {
	propertySimplifiers, err := createPropertySimplifiers(rule)
	if err != nil {
		return nil, err
	}
	return &simplifierImpl{
		propertySimplifiers: propertySimplifiers,
	}, nil
}

// createPropertySimplifiers creates property simplifiers based on the provided rules.
func createPropertySimplifiers(rule *Rule) (map[string]Ruler, error) {
	propertySimplifiers := make(map[string]Ruler)

	for propName, subRule := range rule.PropertySimplifiers {
		propertySimplifier, err := newSimplifier0(subRule)
		if err != nil {
			return nil, err
		}
		propertySimplifiers[propName] = propertySimplifier
	}

	for _, propName := range rule.RemoveProperties {
		propertySimplifiers[propName] = clearRulerIns
	}

	return propertySimplifiers, nil
}

// Simplify applies the rules to the original struct and returns a simplified copy.
func (s *simplifierImpl) Simplify(original interface{}) (interface{}, error) {
	copyValue := reflect.ValueOf(original)
	copyType := reflect.TypeOf(original)

	// Make a deep copy of the original value
	copy := reflect.New(copyType).Elem()
	copy = deepCopy(copy, copyValue)

	// Apply the rules recursively
	s.applyRules("root", copy)

	return copy.Interface(), nil
}

// deepCopy makes a deep copy of the original value recursively.
func deepCopy(copy reflect.Value, original reflect.Value) reflect.Value {
	switch original.Kind() {
	case reflect.Struct:
		for i := 0; i < original.NumField(); i++ {
			copy.Field(i).Set(deepCopy(copy.Field(i), original.Field(i)))
		}
	case reflect.Slice:
		copy.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))
		for i := 0; i < original.Len(); i++ {
			copy.Index(i).Set(deepCopy(copy.Index(i), original.Index(i)))
		}
	case reflect.Ptr:
		copy.Set(reflect.New(original.Type().Elem()))
		copy.Elem().Set(deepCopy(copy.Elem(), original.Elem()))
	default:
		copy.Set(original)
	}

	return copy
}

func (s *clearRuler) applyRules(_ interface{}, field reflect.Value) {
	if field.IsValid() && field.CanSet() {
		field.Set(reflect.Zero(field.Type()))
	}
}

// applyRules applies the rules to the struct recursively.
func (s *simplifierImpl) applyRules(name interface{}, value reflect.Value) {
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return
		}
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.Struct:
		for fieldName, subSimplifier := range s.propertySimplifiers {
			field := value.FieldByName(fieldName)
			if !field.IsValid() || !field.CanInterface() {
				continue
			}

			switch field.Kind() {
			case reflect.Slice:
				for i := 0; i < field.Len(); i++ {
					item := field.Index(i)
					if item.Kind() == reflect.Ptr {
						item = item.Elem()
					}
					subSimplifier.applyRules(fieldName, item)
				}
			default:
				subSimplifier.applyRules(fieldName, field)
			}
		}
	case reflect.Map:
		for _, key := range value.MapKeys() {
			subSimplifier := s.propertySimplifiers[key.String()]
			if subSimplifier != nil {
				subValue := value.MapIndex(key)
				if subValue.Kind() == reflect.Ptr {
					subValue = subValue.Elem()
				}
				if subSimplifier == clearRulerIns {
					value.SetMapIndex(key, reflect.Value{})
				} else {
					subSimplifier.applyRules(key.Interface(), subValue)
				}

			}
		}
	}
}
