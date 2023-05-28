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
//
//                                simplifierImpl
//                                      |
//                      ------------------------------------
//                field1|                            field2|
//                      |                                  |
//                 removeRuler                      simplifierImpl
//                                                         |
//                                                 -------------------
//                                      field1.sub1|       field1.sub2|
//                                                 |                  |
//                                          simplifierImpl      removeRuler
//                             --------------------------------
//                            |                                |
//               field1.sub1.a|                   field1.sub1.b|
//                         removeRuler                  removeRuler
//
type simplifierImpl struct {
	propertySimplifiers map[string]ruler
}

type ruler interface {
	applyRules(parent *reflect.Value, value reflect.Value, valueKey *reflect.Value)
}

// removeRuler for removing a valueKey from parent
type removeRuler struct {
}

var removeRulerSingleton = &removeRuler{}

// NewSimplifier creates a new instance of simplifierImpl with the given rules based on the origin Simplifier.
// If the origin is nil, a new Simplifier is created without any base rules.
func NewSimplifier(rulesJson string) (Simplifier, error) {
	rule := &Rule{}
	err := json.Unmarshal([]byte(rulesJson), rule)
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
func createPropertySimplifiers(rule *Rule) (map[string]ruler, error) {
	propertySimplifiers := make(map[string]ruler)

	for propName, subRule := range rule.PropertySimplifiers {
		propertySimplifier, err := newSimplifier0(subRule)
		if err != nil {
			return nil, err
		}
		propertySimplifiers[propName] = propertySimplifier
	}

	for _, propName := range rule.RemoveProperties {
		propertySimplifiers[propName] = removeRulerSingleton
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
	s.applyRules(nil, copy, nil)

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

func (s *removeRuler) applyRules(parent *reflect.Value, value reflect.Value, valueKey *reflect.Value) {
	if parent == nil {
		return
	}
	p := *parent
	switch p.Kind() {
	case reflect.Struct:
		if value.IsValid() && value.CanSet() {
			value.Set(reflect.Zero(value.Type()))
		}
	case reflect.Map:
		if valueKey == nil {
			return
		}
		p.SetMapIndex(*valueKey, reflect.Value{})
	}
}

// applyRules applies the rules to the struct recursively.
func (s *simplifierImpl) applyRules(parent *reflect.Value, value reflect.Value, valueKey *reflect.Value) {
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
					subSimplifier.applyRules(&value, item, nil)
				}
			default:
				subSimplifier.applyRules(&value, field, nil)
			}
		}
	case reflect.Map:
		for _, key := range value.MapKeys() {
			subSimplifier := s.propertySimplifiers[key.String()]
			if subSimplifier == nil {
				continue
			}
			subValue := value.MapIndex(key)
			if subValue.Kind() == reflect.Ptr {
				if subValue.IsNil() {
					continue
				}
				subValue = subValue.Elem()
			}
			subSimplifier.applyRules(&value, subValue, &key)
		}
	}
}
