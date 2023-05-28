package gosimplifier

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// Rule defines the rule structure for property removal and nested property rules.
type Rule struct {
	RemoveProperties    []string         `json:"remove_properties"`
	PropertySimplifiers map[string]*Rule `json:"property_simplifiers"`
}

// Simplifier defines the interface for struct simplification.
type Simplifier interface {
	// Simplify method:
	// 1. Receives any type of struct or pointer to it, returns the same type of struct(pointer)
	// 2. Will not modify the original, but just make a copy as the return value
	// 3. Removes the properties of the return value according to the rules
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
//                                          simplifierImpl       removeRuler
//                             --------------------------------
//                            |                                |
//               field1.sub1.a|                   field1.sub1.b|
//                         removeRuler                  removeRuler
//
type simplifierImpl struct {
	propertySimplifiers map[string]ruler
	rule                *Rule
}

type ruler interface {
	applyRules(parent *reflect.Value, value reflect.Value, valueKey *reflect.Value)
}

// removeRuler for removing a valueKey from parent
type removeRuler struct {
}

var removeRulerSingleton = &removeRuler{}

// NewSimplifier creates a new instance of simplifierImpl with the given rules
//
// Example:
// {
//  "remove_properties": [ "Test", "Debug" ],
//  "property_simplifiers": {
//    "Data": {
//      "remove_properties": [ "data_test", "data_debug" ]
//    },
//    "EntityList": {
//      "property_simplifiers": {
//        "SubProperties": {
//          "remove_properties": [ "ABC", "DEF" ]
//        }
//      }
//    }
//  }
// }
//
// For struct instance root:
//    var root ExampleStruct
// The following properties will be removed:
//    root.Test
//    root.Debug
//    root.Data.data_test
//    root.Data.data_debug
//    root.EntityList[?].ABC (? could be any index of EntityList)
//    root.EntityList[?].DEF (? could be any index of EntityList)
//
func NewSimplifier(rulesJson string) (Simplifier, error) {
	rule := &Rule{}
	if err := json.Unmarshal([]byte(rulesJson), rule); err != nil {
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
		rule:                rule,
	}, nil
}

// ExtendSimplifier extends the base simplifier with the given rules.
// The new Simplifier will have the rules merge from the base and the given rules.
func ExtendSimplifier(base Simplifier, rulesJson string) (Simplifier, error) {
	baseImpl, ok := base.(*simplifierImpl)
	if !ok {
		return nil, fmt.Errorf("base Simplifier is not the correct type")
	}
	newRule := &Rule{}
	if err := json.Unmarshal([]byte(rulesJson), newRule); err != nil {
		return nil, err
	}
	mergedRule := mergeRules(baseImpl.rule, newRule)
	return newSimplifier0(mergedRule)
}

func mergeRules(rule *Rule, newRule *Rule) *Rule {
	// Copy old rule's remove_properties
	mergedRemoveProperties := make([]string, len(rule.RemoveProperties))
	copy(mergedRemoveProperties, rule.RemoveProperties)

	// Copy old rule's property_simplifiers
	mergedPropertySimplifiers := make(map[string]*Rule)
	for k, v := range rule.PropertySimplifiers {
		mergedPropertySimplifiers[k] = v
	}

	// Merge remove_properties
	for _, prop := range newRule.RemoveProperties {
		if !contains(mergedRemoveProperties, prop) {
			mergedRemoveProperties = append(mergedRemoveProperties, prop)
		}
	}

	// Merge property_simplifiers
	for k, v := range newRule.PropertySimplifiers {
		if _, ok := mergedPropertySimplifiers[k]; ok {
			// If the key already exists, merge the sub-rule
			mergedPropertySimplifiers[k] = mergeRules(mergedPropertySimplifiers[k], v)
		} else {
			// Otherwise, just add the new rule
			mergedPropertySimplifiers[k] = v
		}
	}

	// Return the merged rule
	return &Rule{
		RemoveProperties:    mergedRemoveProperties,
		PropertySimplifiers: mergedPropertySimplifiers,
	}
}

// Helper function to check if a string is in a slice
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
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
	case reflect.Ptr:
		originalValue := original.Elem()
		if !originalValue.IsValid() {
			return original
		}
		newValue := reflect.New(originalValue.Type())
		copy = newValue
		deepCopy(copy.Elem(), originalValue)
	case reflect.Slice:
		copy.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))
		for i := 0; i < original.Len(); i++ {
			deepCopy(copy.Index(i), original.Index(i))
		}
	case reflect.Struct:
		copy.Set(reflect.New(original.Type()).Elem())
		for i := 0; i < original.NumField(); i++ {
			deepCopy(copy.Field(i), original.Field(i))
		}
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
