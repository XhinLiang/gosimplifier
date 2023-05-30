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
//	                 simplifierImpl
//	                       |
//	       ------------------------------------
//	 field1|                            field2|
//	       |                                  |
//	  removeRuler                      simplifierImpl
//	                                          |
//	                                  -------------------
//	                       field1.sub1|       field1.sub2|
//	                                  |                  |
//	                           simplifierImpl       removeRuler
//	              --------------------------------
//	             |                                |
//	field1.sub1.a|                   field1.sub1.b|
//	          removeRuler                  removeRuler
type simplifierImpl struct {
	propertySimplifiers map[string]ruler
	rule                *Rule
}

type ruler interface {
	applyRules(value reflect.Value, mapParent *reflect.Value, mapKey *reflect.Value, root *simplifierImpl)
}

// removeRuler for removing a valueKey from parent
type removeRuler struct {
}

var removeRulerSingleton = &removeRuler{}

// NewSimplifier creates a new instance of simplifierImpl with the given rules
//
// Example:
//
//	{
//	 "remove_properties": [ "field1"],
//	 "property_simplifiers": {
//	   "field2": {
//	     "remove_properties": [ "sub2" ]
//	     "property_simplifiers": {
//	       "sub1": {
//	         "remove_properties": [ "a", "b" ]
//	       }
//	     }
//	   }
//	 }
//	}
//
// For struct instance root:
//
//	var root ExampleStruct
//
// The following properties will be removed:
//
//	root.field1
//	root.field2.sub2
//	root.field2.sub1.a
//	root.field2.sub1.b
//
// Other properties will be kept.
func NewSimplifier(rulesJson string) (Simplifier, error) {
	rule := &Rule{}
	if err := json.Unmarshal([]byte(rulesJson), rule); err != nil {
		return nil, err
	}
	return newSimplifierByRule0(rule)
}

func NewSimplifierByRule(rule *Rule) (Simplifier, error) {
	return newSimplifierByRule0(rule)
}

// newSimplifierByRule0 creates a new instance of simplifierImpl with the given rule
func newSimplifierByRule0(rule *Rule) (*simplifierImpl, error) {
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
	return ExtendSimplifierByRule(baseImpl, newRule)
}

func ExtendSimplifierByRule(baseImpl *simplifierImpl, newRule *Rule) (Simplifier, error) {
	return newSimplifierByRule0(mergeRules(baseImpl.rule, newRule))
}

func mergeRules(rule *Rule, newRule *Rule) *Rule {
	// Copy old rule's remove_properties
	mergedRemoveProperties := make([]string, len(rule.RemoveProperties))
	copy(mergedRemoveProperties, rule.RemoveProperties)

	// Copy old rule's propertySimplifiers
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
		propertySimplifier, err := newSimplifierByRule0(subRule)
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
	cp := reflect.New(copyType).Elem()
	cp = deepCopy(cp, copyValue)

	// Apply the rules recursively
	s.applyRules(cp, nil, nil, s)

	return cp.Interface(), nil
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

func (s *removeRuler) applyRules(value reflect.Value, parent *reflect.Value, mapKey *reflect.Value, rootSimplifier *simplifierImpl) {
	if parent == nil {
		return
	}
	switch p := *parent; p.Kind() {
	case reflect.Struct:
		if value.IsValid() && value.CanSet() {
			value.Set(reflect.Zero(value.Type()))
		}
	case reflect.Map:
		if mapKey == nil {
			return
		}
		p.SetMapIndex(*mapKey, reflect.Value{})
	}
}

func (s *simplifierImpl) applyRules(value reflect.Value, parent *reflect.Value, mapKey *reflect.Value, rootSimpifier *simplifierImpl) {
	s.applyRules0(value, rootSimpifier)
}

func getRealValue(value reflect.Value) reflect.Value {
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return reflect.Value{}
		}
		value = value.Elem()
	}
	return reflect.ValueOf(value.Interface())
}

func (s *simplifierImpl) applyRules0(value reflect.Value, rootSimpifier *simplifierImpl) {
	// applyRules applies the rules to the struct recursively.
	val, kind := value.Interface(), value.Kind()
	if val == nil || kind == reflect.Invalid {
		return
	}
	value = getRealValue(value)
	underlyingKind := value.Kind()

	switch underlyingKind {
	case reflect.Slice:
		for i := 0; i < value.Len(); i++ {
			item := value.Index(i)
			s.applyRules(item, &value, nil, rootSimpifier)
		}
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			field, fieldName := value.Field(i), value.Type().Field(i).Name
			if subSimplifier := s.propertySimplifiers[fieldName]; subSimplifier == nil {
				rootSimpifier.applyRules0(field, rootSimpifier)
			} else {
				subSimplifier.applyRules(field, &value, nil, rootSimpifier)
			}
		}
	case reflect.Map:
		for _, mapKey := range value.MapKeys() {
			mapValue := value.MapIndex(mapKey)
			mapVal, mapKeyStr := mapValue.Interface(), mapKey.String()
			if mapVal == nil && mapKeyStr == "" {
				continue
			}
			if mapValue.IsZero() {
				removeRulerSingleton.applyRules(mapValue, &value, &mapKey, rootSimpifier)
				continue
			}
			if subSimplifier := s.propertySimplifiers[mapKeyStr]; subSimplifier != nil {
				subSimplifier.applyRules(mapValue, &value, &mapKey, rootSimpifier)
				continue
			}
			rootSimpifier.applyRules0(mapValue, rootSimpifier)
		}
	}
}
