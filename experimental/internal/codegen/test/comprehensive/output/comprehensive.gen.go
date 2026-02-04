package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"mime/multipart"
	"regexp"
	"time"
)

// #/components/schemas/AllTypesRequired
type AllTypesRequiredSchemaComponent struct {
	IntField      int       `json:"intField"`
	Int32Field    int32     `json:"int32Field"`
	Int64Field    int64     `json:"int64Field"`
	FloatField    float32   `json:"floatField"`
	DoubleField   float64   `json:"doubleField"`
	NumberField   float32   `json:"numberField"`
	StringField   string    `json:"stringField"`
	BoolField     bool      `json:"boolField"`
	DateField     Date      `json:"dateField"`
	DateTimeField time.Time `json:"dateTimeField"`
	UUIDField     UUID      `json:"uuidField"`
	EmailField    Email     `json:"emailField"`
	URIField      string    `json:"uriField"`
	HostnameField string    `json:"hostnameField"`
	Ipv4Field     string    `json:"ipv4Field"`
	Ipv6Field     string    `json:"ipv6Field"`
	ByteField     []byte    `json:"byteField"`
	BinaryField   File      `json:"binaryField"`
	PasswordField string    `json:"passwordField"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *AllTypesRequiredSchemaComponent) ApplyDefaults() {
}

type AllTypesRequired = AllTypesRequiredSchemaComponent

// #/components/schemas/AllTypesOptional
type AllTypesOptionalSchemaComponent struct {
	IntField      *int       `json:"intField,omitempty"`
	Int32Field    *int32     `json:"int32Field,omitempty"`
	Int64Field    *int64     `json:"int64Field,omitempty"`
	FloatField    *float32   `json:"floatField,omitempty"`
	DoubleField   *float64   `json:"doubleField,omitempty"`
	NumberField   *float32   `json:"numberField,omitempty"`
	StringField   *string    `json:"stringField,omitempty"`
	BoolField     *bool      `json:"boolField,omitempty"`
	DateField     *Date      `json:"dateField,omitempty"`
	DateTimeField *time.Time `json:"dateTimeField,omitempty"`
	UUIDField     *UUID      `json:"uuidField,omitempty"`
	EmailField    *Email     `json:"emailField,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *AllTypesOptionalSchemaComponent) ApplyDefaults() {
}

type AllTypesOptional = AllTypesOptionalSchemaComponent

// #/components/schemas/NullableRequired
type NullableRequiredSchemaComponent struct {
	NullableString *string                         `json:"nullableString"`
	NullableInt    *int                            `json:"nullableInt"`
	NullableObject *NullableRequiredNullableObject `json:"nullableObject"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *NullableRequiredSchemaComponent) ApplyDefaults() {
	if s.NullableObject != nil {
		s.NullableObject.ApplyDefaults()
	}
}

type NullableRequired = NullableRequiredSchemaComponent

// #/components/schemas/NullableRequired/properties/nullableObject
type NullableRequiredNullableObjectPropertySchemaComponent struct {
	Name *string `json:"name,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *NullableRequiredNullableObjectPropertySchemaComponent) ApplyDefaults() {
}

type NullableRequiredNullableObject = NullableRequiredNullableObjectPropertySchemaComponent

// #/components/schemas/NullableOptional
type NullableOptionalSchemaComponent struct {
	NullableString *string `json:"nullableString,omitempty"`
	NullableInt    *int    `json:"nullableInt,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *NullableOptionalSchemaComponent) ApplyDefaults() {
}

type NullableOptional = NullableOptionalSchemaComponent

// #/components/schemas/ArrayTypes
type ArrayTypesSchemaComponent struct {
	StringArray          []string                          `json:"stringArray,omitempty"`
	IntArray             []int                             `json:"intArray,omitempty"`
	ObjectArray          []SimpleObject                    `json:"objectArray,omitempty"`
	InlineObjectArray    []ArrayTypesInlineObjectArrayItem `json:"inlineObjectArray,omitempty"`
	NestedArray          [][]string                        `json:"nestedArray,omitempty"`
	NullableArray        []string                          `json:"nullableArray,omitempty"`
	ArrayWithConstraints []string                          `json:"arrayWithConstraints,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *ArrayTypesSchemaComponent) ApplyDefaults() {
}

type ArrayTypes = ArrayTypesSchemaComponent

// #/components/schemas/ArrayTypes/properties/objectArray
type ArrayTypesObjectArrayPropertySchemaComponent = []SimpleObject

type ArrayTypesObjectArray = ArrayTypesObjectArrayPropertySchemaComponent

// #/components/schemas/ArrayTypes/properties/inlineObjectArray
type ArrayTypesInlineObjectArrayPropertySchemaComponent = []ArrayTypesInlineObjectArrayItem

type ArrayTypesInlineObjectArray = ArrayTypesInlineObjectArrayPropertySchemaComponent

// #/components/schemas/ArrayTypes/properties/inlineObjectArray/items
type ArrayTypesInlineObjectArrayItemPropertySchemaComponent struct {
	ID   *int    `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *ArrayTypesInlineObjectArrayItemPropertySchemaComponent) ApplyDefaults() {
}

type ArrayTypesInlineObjectArrayItem = ArrayTypesInlineObjectArrayItemPropertySchemaComponent

// #/components/schemas/SimpleObject
type SimpleObjectSchemaComponent struct {
	ID   int     `json:"id"`
	Name *string `json:"name,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *SimpleObjectSchemaComponent) ApplyDefaults() {
}

type SimpleObject = SimpleObjectSchemaComponent

// #/components/schemas/NestedObject
type NestedObjectSchemaComponent struct {
	Outer *NestedObjectOuter `json:"outer,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *NestedObjectSchemaComponent) ApplyDefaults() {
	if s.Outer != nil {
		s.Outer.ApplyDefaults()
	}
}

type NestedObject = NestedObjectSchemaComponent

// #/components/schemas/NestedObject/properties/outer
type NestedObjectOuterPropertySchemaComponent struct {
	Inner *NestedObjectOuterInner `json:"inner,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *NestedObjectOuterPropertySchemaComponent) ApplyDefaults() {
	if s.Inner != nil {
		s.Inner.ApplyDefaults()
	}
}

type NestedObjectOuter = NestedObjectOuterPropertySchemaComponent

// #/components/schemas/NestedObject/properties/outer/properties/inner
type NestedObjectOuterInnerPropertyPropertySchemaComponent struct {
	Value *string `json:"value,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *NestedObjectOuterInnerPropertyPropertySchemaComponent) ApplyDefaults() {
}

type NestedObjectOuterInner = NestedObjectOuterInnerPropertyPropertySchemaComponent

// #/components/schemas/AdditionalPropsAny
type AdditionalPropsAnySchemaComponent = map[string]any

type AdditionalPropsAny = AdditionalPropsAnySchemaComponent

// #/components/schemas/AdditionalPropsNone
type AdditionalPropsNoneSchemaComponent struct {
	Known                *string        `json:"known,omitempty"`
	AdditionalProperties map[string]any `json:"-"`
}

func (s AdditionalPropsNoneSchemaComponent) MarshalJSON() ([]byte, error) {
	result := make(map[string]any)

	if s.Known != nil {
		result["known"] = s.Known
	}

	// Add additional properties
	for k, v := range s.AdditionalProperties {
		result[k] = v
	}

	return json.Marshal(result)
}

func (s *AdditionalPropsNoneSchemaComponent) UnmarshalJSON(data []byte) error {
	// Known fields
	knownFields := map[string]bool{
		"known": true,
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if v, ok := raw["known"]; ok {
		var val string
		if err := json.Unmarshal(v, &val); err != nil {
			return err
		}
		s.Known = &val
	}

	// Collect additional properties
	s.AdditionalProperties = make(map[string]any)
	for k, v := range raw {
		if !knownFields[k] {
			var val any
			if err := json.Unmarshal(v, &val); err != nil {
				return err
			}
			s.AdditionalProperties[k] = val
		}
	}

	return nil
}

// ApplyDefaults sets default values for fields that are nil.
func (s *AdditionalPropsNoneSchemaComponent) ApplyDefaults() {
}

type AdditionalPropsNone = AdditionalPropsNoneSchemaComponent

// #/components/schemas/AdditionalPropsTyped
type AdditionalPropsTypedSchemaComponent = map[string]int

type AdditionalPropsTyped = AdditionalPropsTypedSchemaComponent

// #/components/schemas/AdditionalPropsObject
type AdditionalPropsObjectSchemaComponent = map[string]any

type AdditionalPropsObject = AdditionalPropsObjectSchemaComponent

// #/components/schemas/AdditionalPropsWithProps
type AdditionalPropsWithPropsSchemaComponent struct {
	ID                   *int              `json:"id,omitempty"`
	AdditionalProperties map[string]string `json:"-"`
}

func (s AdditionalPropsWithPropsSchemaComponent) MarshalJSON() ([]byte, error) {
	result := make(map[string]any)

	if s.ID != nil {
		result["id"] = s.ID
	}

	// Add additional properties
	for k, v := range s.AdditionalProperties {
		result[k] = v
	}

	return json.Marshal(result)
}

func (s *AdditionalPropsWithPropsSchemaComponent) UnmarshalJSON(data []byte) error {
	// Known fields
	knownFields := map[string]bool{
		"id": true,
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if v, ok := raw["id"]; ok {
		var val int
		if err := json.Unmarshal(v, &val); err != nil {
			return err
		}
		s.ID = &val
	}

	// Collect additional properties
	s.AdditionalProperties = make(map[string]string)
	for k, v := range raw {
		if !knownFields[k] {
			var val string
			if err := json.Unmarshal(v, &val); err != nil {
				return err
			}
			s.AdditionalProperties[k] = val
		}
	}

	return nil
}

// ApplyDefaults sets default values for fields that are nil.
func (s *AdditionalPropsWithPropsSchemaComponent) ApplyDefaults() {
}

type AdditionalPropsWithProps = AdditionalPropsWithPropsSchemaComponent

// #/components/schemas/StringEnum
type StringEnumSchemaComponent string

const (
	StringEnumSchemaComponent_value1 StringEnumSchemaComponent = "value1"
	StringEnumSchemaComponent_value2 StringEnumSchemaComponent = "value2"
	StringEnumSchemaComponent_value3 StringEnumSchemaComponent = "value3"
)

type StringEnum = StringEnumSchemaComponent

// #/components/schemas/IntegerEnum
type IntegerEnumSchemaComponent int

const (
	IntegerEnumSchemaComponent_N1 IntegerEnumSchemaComponent = 1
	IntegerEnumSchemaComponent_N2 IntegerEnumSchemaComponent = 2
	IntegerEnumSchemaComponent_N3 IntegerEnumSchemaComponent = 3
)

type IntegerEnum = IntegerEnumSchemaComponent

// #/components/schemas/ObjectWithEnum
type ObjectWithEnumSchemaComponent struct {
	Status   *string `json:"status,omitempty"`
	Priority *int    `json:"priority,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *ObjectWithEnumSchemaComponent) ApplyDefaults() {
}

type ObjectWithEnum = ObjectWithEnumSchemaComponent

// #/components/schemas/ObjectWithEnum/properties/status
type ObjectWithEnumStatusPropertySchemaComponent string

const (
	ObjectWithEnumStatusPropertySchemaComponent_pending   ObjectWithEnumStatusPropertySchemaComponent = "pending"
	ObjectWithEnumStatusPropertySchemaComponent_active    ObjectWithEnumStatusPropertySchemaComponent = "active"
	ObjectWithEnumStatusPropertySchemaComponent_completed ObjectWithEnumStatusPropertySchemaComponent = "completed"
)

type ObjectWithEnumStatus = ObjectWithEnumStatusPropertySchemaComponent

// #/components/schemas/ObjectWithEnum/properties/priority
type ObjectWithEnumPriorityPropertySchemaComponent int

const (
	ObjectWithEnumPriorityPropertySchemaComponent_N1 ObjectWithEnumPriorityPropertySchemaComponent = 1
	ObjectWithEnumPriorityPropertySchemaComponent_N2 ObjectWithEnumPriorityPropertySchemaComponent = 2
	ObjectWithEnumPriorityPropertySchemaComponent_N3 ObjectWithEnumPriorityPropertySchemaComponent = 3
)

type ObjectWithEnumPriority = ObjectWithEnumPriorityPropertySchemaComponent

// #/components/schemas/InlineEnumInProperty
type InlineEnumInPropertySchemaComponent struct {
	InlineStatus *string `json:"inlineStatus,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *InlineEnumInPropertySchemaComponent) ApplyDefaults() {
}

type InlineEnumInProperty = InlineEnumInPropertySchemaComponent

// #/components/schemas/InlineEnumInProperty/properties/inlineStatus
type InlineEnumInPropertyInlineStatusPropertySchemaComponent string

const (
	InlineEnumInPropertyInlineStatusPropertySchemaComponent_on  InlineEnumInPropertyInlineStatusPropertySchemaComponent = "on"
	InlineEnumInPropertyInlineStatusPropertySchemaComponent_off InlineEnumInPropertyInlineStatusPropertySchemaComponent = "off"
)

type InlineEnumInPropertyInlineStatus = InlineEnumInPropertyInlineStatusPropertySchemaComponent

// #/components/schemas/BaseProperties
type BasePropertiesSchemaComponent struct {
	ID        *int       `json:"id,omitempty"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *BasePropertiesSchemaComponent) ApplyDefaults() {
}

type BaseProperties = BasePropertiesSchemaComponent

// #/components/schemas/ExtendedObject
type ExtendedObjectSchemaComponent struct {
	ID          *int       `json:"id,omitempty"`
	CreatedAt   *time.Time `json:"createdAt,omitempty"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *ExtendedObjectSchemaComponent) ApplyDefaults() {
}

type ExtendedObject = ExtendedObjectSchemaComponent

// #/components/schemas/ExtendedObject/allOf/1
type ExtendedObjectN1AllOfSchemaComponent struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *ExtendedObjectN1AllOfSchemaComponent) ApplyDefaults() {
}

type ExtendedObjectAllOf1 = ExtendedObjectN1AllOfSchemaComponent

// #/components/schemas/DeepInheritance
type DeepInheritanceSchemaComponent struct {
	Extra *string `json:"extra,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *DeepInheritanceSchemaComponent) ApplyDefaults() {
}

type DeepInheritance = DeepInheritanceSchemaComponent

// #/components/schemas/DeepInheritance/allOf/1
type DeepInheritanceN1AllOfSchemaComponent struct {
	Extra *string `json:"extra,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *DeepInheritanceN1AllOfSchemaComponent) ApplyDefaults() {
}

type DeepInheritanceAllOf1 = DeepInheritanceN1AllOfSchemaComponent

// #/components/schemas/AllOfMultipleRefs
type AllOfMultipleRefsSchemaComponent struct {
	ID        *int       `json:"id,omitempty"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	Name      *string    `json:"name,omitempty"`
	Merged    *bool      `json:"merged,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *AllOfMultipleRefsSchemaComponent) ApplyDefaults() {
}

type AllOfMultipleRefs = AllOfMultipleRefsSchemaComponent

// #/components/schemas/AllOfMultipleRefs/allOf/2
type AllOfMultipleRefsN2AllOfSchemaComponent struct {
	Merged *bool `json:"merged,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *AllOfMultipleRefsN2AllOfSchemaComponent) ApplyDefaults() {
}

type AllOfMultipleRefsAllOf2 = AllOfMultipleRefsN2AllOfSchemaComponent

// #/components/schemas/AllOfInlineOnly
type AllOfInlineOnlySchemaComponent struct {
	First  *string `json:"first,omitempty"`
	Second *int    `json:"second,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *AllOfInlineOnlySchemaComponent) ApplyDefaults() {
}

type AllOfInlineOnly = AllOfInlineOnlySchemaComponent

// #/components/schemas/AllOfInlineOnly/allOf/0
type AllOfInlineOnlyN0AllOfSchemaComponent struct {
	First *string `json:"first,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *AllOfInlineOnlyN0AllOfSchemaComponent) ApplyDefaults() {
}

type AllOfInlineOnlyAllOf0 = AllOfInlineOnlyN0AllOfSchemaComponent

// #/components/schemas/AllOfInlineOnly/allOf/1
type AllOfInlineOnlyN1AllOfSchemaComponent struct {
	Second *int `json:"second,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *AllOfInlineOnlyN1AllOfSchemaComponent) ApplyDefaults() {
}

type AllOfInlineOnlyAllOf1 = AllOfInlineOnlyN1AllOfSchemaComponent

// #/components/schemas/AnyOfPrimitives
type AnyOfPrimitivesSchemaComponent struct {
	_String0 *string
	_Int1    *int
}

func (u AnyOfPrimitivesSchemaComponent) MarshalJSON() ([]byte, error) {
	if u._String0 != nil {
		return json.Marshal(u._String0)
	}
	if u._Int1 != nil {
		return json.Marshal(u._Int1)
	}
	return []byte("null"), nil
}

func (u *AnyOfPrimitivesSchemaComponent) UnmarshalJSON(data []byte) error {
	var v0 string
	if err := json.Unmarshal(data, &v0); err == nil {
		u._String0 = &v0
	}

	var v1 int
	if err := json.Unmarshal(data, &v1); err == nil {
		u._Int1 = &v1
	}

	return nil
}

// ApplyDefaults sets default values for fields that are nil.
func (u *AnyOfPrimitivesSchemaComponent) ApplyDefaults() {
}

type AnyOfPrimitives = AnyOfPrimitivesSchemaComponent

// #/components/schemas/AnyOfObjects
type AnyOfObjectsSchemaComponent struct {
	SimpleObject   *SimpleObject
	BaseProperties *BaseProperties
}

func (u AnyOfObjectsSchemaComponent) MarshalJSON() ([]byte, error) {
	result := make(map[string]any)

	if u.SimpleObject != nil {
		data, err := json.Marshal(u.SimpleObject)
		if err != nil {
			return nil, err
		}
		var m map[string]any
		if err := json.Unmarshal(data, &m); err == nil {
			for k, v := range m {
				result[k] = v
			}
		}
	}
	if u.BaseProperties != nil {
		data, err := json.Marshal(u.BaseProperties)
		if err != nil {
			return nil, err
		}
		var m map[string]any
		if err := json.Unmarshal(data, &m); err == nil {
			for k, v := range m {
				result[k] = v
			}
		}
	}

	return json.Marshal(result)
}

func (u *AnyOfObjectsSchemaComponent) UnmarshalJSON(data []byte) error {
	var v0 SimpleObject
	if err := json.Unmarshal(data, &v0); err == nil {
		u.SimpleObject = &v0
	}

	var v1 BaseProperties
	if err := json.Unmarshal(data, &v1); err == nil {
		u.BaseProperties = &v1
	}

	return nil
}

// ApplyDefaults sets default values for fields that are nil.
func (u *AnyOfObjectsSchemaComponent) ApplyDefaults() {
	if u.SimpleObject != nil {
		u.SimpleObject.ApplyDefaults()
	}
	if u.BaseProperties != nil {
		u.BaseProperties.ApplyDefaults()
	}
}

type AnyOfObjects = AnyOfObjectsSchemaComponent

// #/components/schemas/AnyOfMixed
type AnyOfMixedSchemaComponent struct {
	_String0         *string
	SimpleObject     *SimpleObject
	AnyOfMixedAnyOf2 *AnyOfMixedAnyOf2
}

func (u AnyOfMixedSchemaComponent) MarshalJSON() ([]byte, error) {
	result := make(map[string]any)

	if u._String0 != nil {
		return json.Marshal(u._String0)
	}
	if u.SimpleObject != nil {
		data, err := json.Marshal(u.SimpleObject)
		if err != nil {
			return nil, err
		}
		var m map[string]any
		if err := json.Unmarshal(data, &m); err == nil {
			for k, v := range m {
				result[k] = v
			}
		}
	}
	if u.AnyOfMixedAnyOf2 != nil {
		data, err := json.Marshal(u.AnyOfMixedAnyOf2)
		if err != nil {
			return nil, err
		}
		var m map[string]any
		if err := json.Unmarshal(data, &m); err == nil {
			for k, v := range m {
				result[k] = v
			}
		}
	}

	return json.Marshal(result)
}

func (u *AnyOfMixedSchemaComponent) UnmarshalJSON(data []byte) error {
	var v0 string
	if err := json.Unmarshal(data, &v0); err == nil {
		u._String0 = &v0
	}

	var v1 SimpleObject
	if err := json.Unmarshal(data, &v1); err == nil {
		u.SimpleObject = &v1
	}

	var v2 AnyOfMixedAnyOf2
	if err := json.Unmarshal(data, &v2); err == nil {
		u.AnyOfMixedAnyOf2 = &v2
	}

	return nil
}

// ApplyDefaults sets default values for fields that are nil.
func (u *AnyOfMixedSchemaComponent) ApplyDefaults() {
	if u.SimpleObject != nil {
		u.SimpleObject.ApplyDefaults()
	}
	if u.AnyOfMixedAnyOf2 != nil {
		u.AnyOfMixedAnyOf2.ApplyDefaults()
	}
}

type AnyOfMixed = AnyOfMixedSchemaComponent

// #/components/schemas/AnyOfMixed/anyOf/2
type AnyOfMixedN2AnyOfSchemaComponent struct {
	Inline *bool `json:"inline,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *AnyOfMixedN2AnyOfSchemaComponent) ApplyDefaults() {
}

type AnyOfMixedAnyOf2 = AnyOfMixedN2AnyOfSchemaComponent

// #/components/schemas/AnyOfNullable
type AnyOfNullableSchemaComponent struct {
	_String0 *string
	_Any1    *any
}

func (u AnyOfNullableSchemaComponent) MarshalJSON() ([]byte, error) {
	if u._String0 != nil {
		return json.Marshal(u._String0)
	}
	if u._Any1 != nil {
		return json.Marshal(u._Any1)
	}
	return []byte("null"), nil
}

func (u *AnyOfNullableSchemaComponent) UnmarshalJSON(data []byte) error {
	var v0 string
	if err := json.Unmarshal(data, &v0); err == nil {
		u._String0 = &v0
	}

	var v1 any
	if err := json.Unmarshal(data, &v1); err == nil {
		u._Any1 = &v1
	}

	return nil
}

// ApplyDefaults sets default values for fields that are nil.
func (u *AnyOfNullableSchemaComponent) ApplyDefaults() {
}

type AnyOfNullable = AnyOfNullableSchemaComponent

// #/components/schemas/ObjectWithAnyOfProperty
type ObjectWithAnyOfPropertySchemaComponent struct {
	Value *ObjectWithAnyOfPropertyValue `json:"value,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *ObjectWithAnyOfPropertySchemaComponent) ApplyDefaults() {
}

type ObjectWithAnyOfProperty = ObjectWithAnyOfPropertySchemaComponent

// #/components/schemas/ObjectWithAnyOfProperty/properties/value
type ObjectWithAnyOfPropertyValuePropertySchemaComponent struct {
	_String0 *string
	_Int1    *int
	_Bool2   *bool
}

func (u ObjectWithAnyOfPropertyValuePropertySchemaComponent) MarshalJSON() ([]byte, error) {
	if u._String0 != nil {
		return json.Marshal(u._String0)
	}
	if u._Int1 != nil {
		return json.Marshal(u._Int1)
	}
	if u._Bool2 != nil {
		return json.Marshal(u._Bool2)
	}
	return []byte("null"), nil
}

func (u *ObjectWithAnyOfPropertyValuePropertySchemaComponent) UnmarshalJSON(data []byte) error {
	var v0 string
	if err := json.Unmarshal(data, &v0); err == nil {
		u._String0 = &v0
	}

	var v1 int
	if err := json.Unmarshal(data, &v1); err == nil {
		u._Int1 = &v1
	}

	var v2 bool
	if err := json.Unmarshal(data, &v2); err == nil {
		u._Bool2 = &v2
	}

	return nil
}

// ApplyDefaults sets default values for fields that are nil.
func (u *ObjectWithAnyOfPropertyValuePropertySchemaComponent) ApplyDefaults() {
}

type ObjectWithAnyOfPropertyValue = ObjectWithAnyOfPropertyValuePropertySchemaComponent

// #/components/schemas/ArrayOfAnyOf
type ArrayOfAnyOfSchemaComponent = []ArrayOfAnyOfItem

type ArrayOfAnyOf = ArrayOfAnyOfSchemaComponent

// #/components/schemas/ArrayOfAnyOf/items
type ArrayOfAnyOfItemSchemaComponent struct {
	SimpleObject   *SimpleObject
	BaseProperties *BaseProperties
}

func (u ArrayOfAnyOfItemSchemaComponent) MarshalJSON() ([]byte, error) {
	result := make(map[string]any)

	if u.SimpleObject != nil {
		data, err := json.Marshal(u.SimpleObject)
		if err != nil {
			return nil, err
		}
		var m map[string]any
		if err := json.Unmarshal(data, &m); err == nil {
			for k, v := range m {
				result[k] = v
			}
		}
	}
	if u.BaseProperties != nil {
		data, err := json.Marshal(u.BaseProperties)
		if err != nil {
			return nil, err
		}
		var m map[string]any
		if err := json.Unmarshal(data, &m); err == nil {
			for k, v := range m {
				result[k] = v
			}
		}
	}

	return json.Marshal(result)
}

func (u *ArrayOfAnyOfItemSchemaComponent) UnmarshalJSON(data []byte) error {
	var v0 SimpleObject
	if err := json.Unmarshal(data, &v0); err == nil {
		u.SimpleObject = &v0
	}

	var v1 BaseProperties
	if err := json.Unmarshal(data, &v1); err == nil {
		u.BaseProperties = &v1
	}

	return nil
}

// ApplyDefaults sets default values for fields that are nil.
func (u *ArrayOfAnyOfItemSchemaComponent) ApplyDefaults() {
	if u.SimpleObject != nil {
		u.SimpleObject.ApplyDefaults()
	}
	if u.BaseProperties != nil {
		u.BaseProperties.ApplyDefaults()
	}
}

type ArrayOfAnyOfItem = ArrayOfAnyOfItemSchemaComponent

// #/components/schemas/OneOfSimple
type OneOfSimpleSchemaComponent struct {
	SimpleObject   *SimpleObject
	BaseProperties *BaseProperties
}

func (u OneOfSimpleSchemaComponent) MarshalJSON() ([]byte, error) {
	var count int
	var data []byte
	var err error

	if u.SimpleObject != nil {
		count++
		data, err = json.Marshal(u.SimpleObject)
		if err != nil {
			return nil, err
		}
	}
	if u.BaseProperties != nil {
		count++
		data, err = json.Marshal(u.BaseProperties)
		if err != nil {
			return nil, err
		}
	}

	if count != 1 {
		return nil, fmt.Errorf("OneOfSimpleSchemaComponent: exactly one member must be set, got %d", count)
	}

	return data, nil
}

func (u *OneOfSimpleSchemaComponent) UnmarshalJSON(data []byte) error {
	var successCount int

	var v0 SimpleObject
	if err := json.Unmarshal(data, &v0); err == nil {
		u.SimpleObject = &v0
		successCount++
	}

	var v1 BaseProperties
	if err := json.Unmarshal(data, &v1); err == nil {
		u.BaseProperties = &v1
		successCount++
	}

	if successCount != 1 {
		return fmt.Errorf("OneOfSimpleSchemaComponent: expected exactly one type to match, got %d", successCount)
	}

	return nil
}

// ApplyDefaults sets default values for fields that are nil.
func (u *OneOfSimpleSchemaComponent) ApplyDefaults() {
	if u.SimpleObject != nil {
		u.SimpleObject.ApplyDefaults()
	}
	if u.BaseProperties != nil {
		u.BaseProperties.ApplyDefaults()
	}
}

type OneOfSimple = OneOfSimpleSchemaComponent

// #/components/schemas/OneOfWithDiscriminator
type OneOfWithDiscriminatorSchemaComponent struct {
	Cat *Cat
	Dog *Dog
}

func (u OneOfWithDiscriminatorSchemaComponent) MarshalJSON() ([]byte, error) {
	var count int
	var data []byte
	var err error

	if u.Cat != nil {
		count++
		data, err = json.Marshal(u.Cat)
		if err != nil {
			return nil, err
		}
	}
	if u.Dog != nil {
		count++
		data, err = json.Marshal(u.Dog)
		if err != nil {
			return nil, err
		}
	}

	if count != 1 {
		return nil, fmt.Errorf("OneOfWithDiscriminatorSchemaComponent: exactly one member must be set, got %d", count)
	}

	return data, nil
}

func (u *OneOfWithDiscriminatorSchemaComponent) UnmarshalJSON(data []byte) error {
	var successCount int

	var v0 Cat
	if err := json.Unmarshal(data, &v0); err == nil {
		u.Cat = &v0
		successCount++
	}

	var v1 Dog
	if err := json.Unmarshal(data, &v1); err == nil {
		u.Dog = &v1
		successCount++
	}

	if successCount != 1 {
		return fmt.Errorf("OneOfWithDiscriminatorSchemaComponent: expected exactly one type to match, got %d", successCount)
	}

	return nil
}

// ApplyDefaults sets default values for fields that are nil.
func (u *OneOfWithDiscriminatorSchemaComponent) ApplyDefaults() {
	if u.Cat != nil {
		u.Cat.ApplyDefaults()
	}
	if u.Dog != nil {
		u.Dog.ApplyDefaults()
	}
}

type OneOfWithDiscriminator = OneOfWithDiscriminatorSchemaComponent

// #/components/schemas/OneOfWithDiscriminatorMapping
type OneOfWithDiscriminatorMappingSchemaComponent struct {
	Cat *Cat
	Dog *Dog
}

func (u OneOfWithDiscriminatorMappingSchemaComponent) MarshalJSON() ([]byte, error) {
	var count int
	var data []byte
	var err error

	if u.Cat != nil {
		count++
		data, err = json.Marshal(u.Cat)
		if err != nil {
			return nil, err
		}
	}
	if u.Dog != nil {
		count++
		data, err = json.Marshal(u.Dog)
		if err != nil {
			return nil, err
		}
	}

	if count != 1 {
		return nil, fmt.Errorf("OneOfWithDiscriminatorMappingSchemaComponent: exactly one member must be set, got %d", count)
	}

	return data, nil
}

func (u *OneOfWithDiscriminatorMappingSchemaComponent) UnmarshalJSON(data []byte) error {
	var successCount int

	var v0 Cat
	if err := json.Unmarshal(data, &v0); err == nil {
		u.Cat = &v0
		successCount++
	}

	var v1 Dog
	if err := json.Unmarshal(data, &v1); err == nil {
		u.Dog = &v1
		successCount++
	}

	if successCount != 1 {
		return fmt.Errorf("OneOfWithDiscriminatorMappingSchemaComponent: expected exactly one type to match, got %d", successCount)
	}

	return nil
}

// ApplyDefaults sets default values for fields that are nil.
func (u *OneOfWithDiscriminatorMappingSchemaComponent) ApplyDefaults() {
	if u.Cat != nil {
		u.Cat.ApplyDefaults()
	}
	if u.Dog != nil {
		u.Dog.ApplyDefaults()
	}
}

type OneOfWithDiscriminatorMapping = OneOfWithDiscriminatorMappingSchemaComponent

// #/components/schemas/Cat
type CatSchemaComponent struct {
	PetType       string   `json:"petType"`
	Meow          string   `json:"meow"`
	WhiskerLength *float32 `json:"whiskerLength,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *CatSchemaComponent) ApplyDefaults() {
}

type Cat = CatSchemaComponent

// #/components/schemas/Dog
type DogSchemaComponent struct {
	PetType    string   `json:"petType"`
	Bark       string   `json:"bark"`
	TailLength *float32 `json:"tailLength,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *DogSchemaComponent) ApplyDefaults() {
}

type Dog = DogSchemaComponent

// #/components/schemas/OneOfInline
type OneOfInlineSchemaComponent struct {
	OneOfInlineOneOf0 *OneOfInlineOneOf0
	OneOfInlineOneOf1 *OneOfInlineOneOf1
}

func (u OneOfInlineSchemaComponent) MarshalJSON() ([]byte, error) {
	var count int
	var data []byte
	var err error

	if u.OneOfInlineOneOf0 != nil {
		count++
		data, err = json.Marshal(u.OneOfInlineOneOf0)
		if err != nil {
			return nil, err
		}
	}
	if u.OneOfInlineOneOf1 != nil {
		count++
		data, err = json.Marshal(u.OneOfInlineOneOf1)
		if err != nil {
			return nil, err
		}
	}

	if count != 1 {
		return nil, fmt.Errorf("OneOfInlineSchemaComponent: exactly one member must be set, got %d", count)
	}

	return data, nil
}

func (u *OneOfInlineSchemaComponent) UnmarshalJSON(data []byte) error {
	var successCount int

	var v0 OneOfInlineOneOf0
	if err := json.Unmarshal(data, &v0); err == nil {
		u.OneOfInlineOneOf0 = &v0
		successCount++
	}

	var v1 OneOfInlineOneOf1
	if err := json.Unmarshal(data, &v1); err == nil {
		u.OneOfInlineOneOf1 = &v1
		successCount++
	}

	if successCount != 1 {
		return fmt.Errorf("OneOfInlineSchemaComponent: expected exactly one type to match, got %d", successCount)
	}

	return nil
}

// ApplyDefaults sets default values for fields that are nil.
func (u *OneOfInlineSchemaComponent) ApplyDefaults() {
	if u.OneOfInlineOneOf0 != nil {
		u.OneOfInlineOneOf0.ApplyDefaults()
	}
	if u.OneOfInlineOneOf1 != nil {
		u.OneOfInlineOneOf1.ApplyDefaults()
	}
}

type OneOfInline = OneOfInlineSchemaComponent

// #/components/schemas/OneOfInline/oneOf/0
type OneOfInlineN0OneOfSchemaComponent struct {
	OptionA *string `json:"optionA,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *OneOfInlineN0OneOfSchemaComponent) ApplyDefaults() {
}

type OneOfInlineOneOf0 = OneOfInlineN0OneOfSchemaComponent

// #/components/schemas/OneOfInline/oneOf/1
type OneOfInlineN1OneOfSchemaComponent struct {
	OptionB *int `json:"optionB,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *OneOfInlineN1OneOfSchemaComponent) ApplyDefaults() {
}

type OneOfInlineOneOf1 = OneOfInlineN1OneOfSchemaComponent

// #/components/schemas/OneOfPrimitives
type OneOfPrimitivesSchemaComponent struct {
	_String0  *string
	_Float321 *float32
	_Bool2    *bool
}

func (u OneOfPrimitivesSchemaComponent) MarshalJSON() ([]byte, error) {
	var count int
	var data []byte
	var err error

	if u._String0 != nil {
		count++
		data, err = json.Marshal(u._String0)
		if err != nil {
			return nil, err
		}
	}
	if u._Float321 != nil {
		count++
		data, err = json.Marshal(u._Float321)
		if err != nil {
			return nil, err
		}
	}
	if u._Bool2 != nil {
		count++
		data, err = json.Marshal(u._Bool2)
		if err != nil {
			return nil, err
		}
	}

	if count != 1 {
		return nil, fmt.Errorf("OneOfPrimitivesSchemaComponent: exactly one member must be set, got %d", count)
	}

	return data, nil
}

func (u *OneOfPrimitivesSchemaComponent) UnmarshalJSON(data []byte) error {
	var successCount int

	var v0 string
	if err := json.Unmarshal(data, &v0); err == nil {
		u._String0 = &v0
		successCount++
	}

	var v1 float32
	if err := json.Unmarshal(data, &v1); err == nil {
		u._Float321 = &v1
		successCount++
	}

	var v2 bool
	if err := json.Unmarshal(data, &v2); err == nil {
		u._Bool2 = &v2
		successCount++
	}

	if successCount != 1 {
		return fmt.Errorf("OneOfPrimitivesSchemaComponent: expected exactly one type to match, got %d", successCount)
	}

	return nil
}

// ApplyDefaults sets default values for fields that are nil.
func (u *OneOfPrimitivesSchemaComponent) ApplyDefaults() {
}

type OneOfPrimitives = OneOfPrimitivesSchemaComponent

// #/components/schemas/ObjectWithOneOfProperty
type ObjectWithOneOfPropertySchemaComponent struct {
	ID      *int                            `json:"id,omitempty"`
	Variant *ObjectWithOneOfPropertyVariant `json:"variant,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *ObjectWithOneOfPropertySchemaComponent) ApplyDefaults() {
}

type ObjectWithOneOfProperty = ObjectWithOneOfPropertySchemaComponent

// #/components/schemas/ObjectWithOneOfProperty/properties/variant
type ObjectWithOneOfPropertyVariantPropertySchemaComponent struct {
	Cat *Cat
	Dog *Dog
}

func (u ObjectWithOneOfPropertyVariantPropertySchemaComponent) MarshalJSON() ([]byte, error) {
	var count int
	var data []byte
	var err error

	if u.Cat != nil {
		count++
		data, err = json.Marshal(u.Cat)
		if err != nil {
			return nil, err
		}
	}
	if u.Dog != nil {
		count++
		data, err = json.Marshal(u.Dog)
		if err != nil {
			return nil, err
		}
	}

	if count != 1 {
		return nil, fmt.Errorf("ObjectWithOneOfPropertyVariantPropertySchemaComponent: exactly one member must be set, got %d", count)
	}

	return data, nil
}

func (u *ObjectWithOneOfPropertyVariantPropertySchemaComponent) UnmarshalJSON(data []byte) error {
	var successCount int

	var v0 Cat
	if err := json.Unmarshal(data, &v0); err == nil {
		u.Cat = &v0
		successCount++
	}

	var v1 Dog
	if err := json.Unmarshal(data, &v1); err == nil {
		u.Dog = &v1
		successCount++
	}

	if successCount != 1 {
		return fmt.Errorf("ObjectWithOneOfPropertyVariantPropertySchemaComponent: expected exactly one type to match, got %d", successCount)
	}

	return nil
}

// ApplyDefaults sets default values for fields that are nil.
func (u *ObjectWithOneOfPropertyVariantPropertySchemaComponent) ApplyDefaults() {
	if u.Cat != nil {
		u.Cat.ApplyDefaults()
	}
	if u.Dog != nil {
		u.Dog.ApplyDefaults()
	}
}

type ObjectWithOneOfPropertyVariant = ObjectWithOneOfPropertyVariantPropertySchemaComponent

// #/components/schemas/AllOfWithOneOf
type AllOfWithOneOfSchemaComponent struct {
	ID                   *int                  `json:"id,omitempty"`
	CreatedAt            *time.Time            `json:"createdAt,omitempty"`
	AllOfWithOneOfAllOf1 *AllOfWithOneOfAllOf1 `json:"-"`
}

func (s AllOfWithOneOfSchemaComponent) MarshalJSON() ([]byte, error) {
	result := make(map[string]any)

	if s.ID != nil {
		result["id"] = s.ID
	}
	if s.CreatedAt != nil {
		result["createdAt"] = s.CreatedAt
	}

	if s.AllOfWithOneOfAllOf1 != nil {
		unionData, err := json.Marshal(s.AllOfWithOneOfAllOf1)
		if err != nil {
			return nil, err
		}
		var unionMap map[string]any
		if err := json.Unmarshal(unionData, &unionMap); err == nil {
			for k, v := range unionMap {
				result[k] = v
			}
		}
	}

	return json.Marshal(result)
}

func (s *AllOfWithOneOfSchemaComponent) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if v, ok := raw["id"]; ok {
		var val int
		if err := json.Unmarshal(v, &val); err != nil {
			return err
		}
		s.ID = &val
	}
	if v, ok := raw["createdAt"]; ok {
		var val time.Time
		if err := json.Unmarshal(v, &val); err != nil {
			return err
		}
		s.CreatedAt = &val
	}

	var AllOfWithOneOfAllOf1Val AllOfWithOneOfAllOf1
	if err := json.Unmarshal(data, &AllOfWithOneOfAllOf1Val); err != nil {
		return err
	}
	s.AllOfWithOneOfAllOf1 = &AllOfWithOneOfAllOf1Val

	return nil
}

// ApplyDefaults sets default values for fields that are nil.
func (s *AllOfWithOneOfSchemaComponent) ApplyDefaults() {
}

type AllOfWithOneOf = AllOfWithOneOfSchemaComponent

// #/components/schemas/AllOfWithOneOf/allOf/1
type AllOfWithOneOfN1AllOfSchemaComponent struct {
	Cat *Cat
	Dog *Dog
}

func (u AllOfWithOneOfN1AllOfSchemaComponent) MarshalJSON() ([]byte, error) {
	var count int
	var data []byte
	var err error

	if u.Cat != nil {
		count++
		data, err = json.Marshal(u.Cat)
		if err != nil {
			return nil, err
		}
	}
	if u.Dog != nil {
		count++
		data, err = json.Marshal(u.Dog)
		if err != nil {
			return nil, err
		}
	}

	if count != 1 {
		return nil, fmt.Errorf("AllOfWithOneOfN1AllOfSchemaComponent: exactly one member must be set, got %d", count)
	}

	return data, nil
}

func (u *AllOfWithOneOfN1AllOfSchemaComponent) UnmarshalJSON(data []byte) error {
	var successCount int

	var v0 Cat
	if err := json.Unmarshal(data, &v0); err == nil {
		u.Cat = &v0
		successCount++
	}

	var v1 Dog
	if err := json.Unmarshal(data, &v1); err == nil {
		u.Dog = &v1
		successCount++
	}

	if successCount != 1 {
		return fmt.Errorf("AllOfWithOneOfN1AllOfSchemaComponent: expected exactly one type to match, got %d", successCount)
	}

	return nil
}

// ApplyDefaults sets default values for fields that are nil.
func (u *AllOfWithOneOfN1AllOfSchemaComponent) ApplyDefaults() {
	if u.Cat != nil {
		u.Cat.ApplyDefaults()
	}
	if u.Dog != nil {
		u.Dog.ApplyDefaults()
	}
}

type AllOfWithOneOfAllOf1 = AllOfWithOneOfN1AllOfSchemaComponent

// #/components/schemas/OneOfWithAllOf
type OneOfWithAllOfSchemaComponent struct {
	OneOfWithAllOfOneOf0 *OneOfWithAllOfOneOf0
	OneOfWithAllOfOneOf1 *OneOfWithAllOfOneOf1
}

func (u OneOfWithAllOfSchemaComponent) MarshalJSON() ([]byte, error) {
	var count int
	var data []byte
	var err error

	if u.OneOfWithAllOfOneOf0 != nil {
		count++
		data, err = json.Marshal(u.OneOfWithAllOfOneOf0)
		if err != nil {
			return nil, err
		}
	}
	if u.OneOfWithAllOfOneOf1 != nil {
		count++
		data, err = json.Marshal(u.OneOfWithAllOfOneOf1)
		if err != nil {
			return nil, err
		}
	}

	if count != 1 {
		return nil, fmt.Errorf("OneOfWithAllOfSchemaComponent: exactly one member must be set, got %d", count)
	}

	return data, nil
}

func (u *OneOfWithAllOfSchemaComponent) UnmarshalJSON(data []byte) error {
	var successCount int

	var v0 OneOfWithAllOfOneOf0
	if err := json.Unmarshal(data, &v0); err == nil {
		u.OneOfWithAllOfOneOf0 = &v0
		successCount++
	}

	var v1 OneOfWithAllOfOneOf1
	if err := json.Unmarshal(data, &v1); err == nil {
		u.OneOfWithAllOfOneOf1 = &v1
		successCount++
	}

	if successCount != 1 {
		return fmt.Errorf("OneOfWithAllOfSchemaComponent: expected exactly one type to match, got %d", successCount)
	}

	return nil
}

// ApplyDefaults sets default values for fields that are nil.
func (u *OneOfWithAllOfSchemaComponent) ApplyDefaults() {
	if u.OneOfWithAllOfOneOf0 != nil {
		u.OneOfWithAllOfOneOf0.ApplyDefaults()
	}
	if u.OneOfWithAllOfOneOf1 != nil {
		u.OneOfWithAllOfOneOf1.ApplyDefaults()
	}
}

type OneOfWithAllOf = OneOfWithAllOfSchemaComponent

// #/components/schemas/OneOfWithAllOf/oneOf/0
type OneOfWithAllOfN0OneOfSchemaComponent struct {
	ID        *int       `json:"id,omitempty"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	Variant   *string    `json:"variant,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *OneOfWithAllOfN0OneOfSchemaComponent) ApplyDefaults() {
}

type OneOfWithAllOfOneOf0 = OneOfWithAllOfN0OneOfSchemaComponent

// #/components/schemas/OneOfWithAllOf/oneOf/0/allOf/1
type OneOfWithAllOfN0N1AllOfOneOfSchemaComponent struct {
	Variant *string `json:"variant,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *OneOfWithAllOfN0N1AllOfOneOfSchemaComponent) ApplyDefaults() {
}

type OneOfWithAllOfOneOf0AllOf1 = OneOfWithAllOfN0N1AllOfOneOfSchemaComponent

// #/components/schemas/OneOfWithAllOf/oneOf/1
type OneOfWithAllOfN1OneOfSchemaComponent struct {
	ID        *int       `json:"id,omitempty"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	Variant   *string    `json:"variant,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *OneOfWithAllOfN1OneOfSchemaComponent) ApplyDefaults() {
}

type OneOfWithAllOfOneOf1 = OneOfWithAllOfN1OneOfSchemaComponent

// #/components/schemas/OneOfWithAllOf/oneOf/1/allOf/1
type OneOfWithAllOfN1N1AllOfOneOfSchemaComponent struct {
	Variant *string `json:"variant,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *OneOfWithAllOfN1N1AllOfOneOfSchemaComponent) ApplyDefaults() {
}

type OneOfWithAllOfOneOf1AllOf1 = OneOfWithAllOfN1N1AllOfOneOfSchemaComponent

// #/components/schemas/TreeNode
type TreeNodeSchemaComponent struct {
	Value    *string    `json:"value,omitempty"`
	Children []TreeNode `json:"children,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *TreeNodeSchemaComponent) ApplyDefaults() {
}

type TreeNode = TreeNodeSchemaComponent

// #/components/schemas/TreeNode/properties/children
type TreeNodeChildrenPropertySchemaComponent = []TreeNode

type TreeNodeChildren = TreeNodeChildrenPropertySchemaComponent

// #/components/schemas/LinkedListNode
type LinkedListNodeSchemaComponent struct {
	Value *int            `json:"value,omitempty"`
	Next  *LinkedListNode `json:"next,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *LinkedListNodeSchemaComponent) ApplyDefaults() {
	if s.Next != nil {
		s.Next.ApplyDefaults()
	}
}

type LinkedListNode = LinkedListNodeSchemaComponent

// #/components/schemas/RecursiveOneOf
type RecursiveOneOfSchemaComponent struct {
	_String0             *string
	RecursiveOneOfOneOf1 *RecursiveOneOfOneOf1
}

func (u RecursiveOneOfSchemaComponent) MarshalJSON() ([]byte, error) {
	var count int
	var data []byte
	var err error

	if u._String0 != nil {
		count++
		data, err = json.Marshal(u._String0)
		if err != nil {
			return nil, err
		}
	}
	if u.RecursiveOneOfOneOf1 != nil {
		count++
		data, err = json.Marshal(u.RecursiveOneOfOneOf1)
		if err != nil {
			return nil, err
		}
	}

	if count != 1 {
		return nil, fmt.Errorf("RecursiveOneOfSchemaComponent: exactly one member must be set, got %d", count)
	}

	return data, nil
}

func (u *RecursiveOneOfSchemaComponent) UnmarshalJSON(data []byte) error {
	var successCount int

	var v0 string
	if err := json.Unmarshal(data, &v0); err == nil {
		u._String0 = &v0
		successCount++
	}

	var v1 RecursiveOneOfOneOf1
	if err := json.Unmarshal(data, &v1); err == nil {
		u.RecursiveOneOfOneOf1 = &v1
		successCount++
	}

	if successCount != 1 {
		return fmt.Errorf("RecursiveOneOfSchemaComponent: expected exactly one type to match, got %d", successCount)
	}

	return nil
}

// ApplyDefaults sets default values for fields that are nil.
func (u *RecursiveOneOfSchemaComponent) ApplyDefaults() {
	if u.RecursiveOneOfOneOf1 != nil {
		u.RecursiveOneOfOneOf1.ApplyDefaults()
	}
}

type RecursiveOneOf = RecursiveOneOfSchemaComponent

// #/components/schemas/RecursiveOneOf/oneOf/1
type RecursiveOneOfN1OneOfSchemaComponent struct {
	Nested *RecursiveOneOf `json:"nested,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *RecursiveOneOfN1OneOfSchemaComponent) ApplyDefaults() {
	if s.Nested != nil {
		s.Nested.ApplyDefaults()
	}
}

type RecursiveOneOfOneOf1 = RecursiveOneOfN1OneOfSchemaComponent

// #/components/schemas/ReadWriteOnly
type ReadWriteOnlySchemaComponent struct {
	ID       int     `json:"id"`
	Password string  `json:"password"`
	Name     *string `json:"name,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *ReadWriteOnlySchemaComponent) ApplyDefaults() {
}

type ReadWriteOnly = ReadWriteOnlySchemaComponent

// #/components/schemas/WithDefaults
type WithDefaultsSchemaComponent struct {
	StringWithDefault *string  `json:"stringWithDefault,omitempty"`
	IntWithDefault    *int     `json:"intWithDefault,omitempty"`
	BoolWithDefault   *bool    `json:"boolWithDefault,omitempty"`
	ArrayWithDefault  []string `json:"arrayWithDefault,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *WithDefaultsSchemaComponent) ApplyDefaults() {
	if s.StringWithDefault == nil {
		v := "default_value"
		s.StringWithDefault = &v
	}
	if s.IntWithDefault == nil {
		v := 42
		s.IntWithDefault = &v
	}
	if s.BoolWithDefault == nil {
		v := true
		s.BoolWithDefault = &v
	}
}

type WithDefaults = WithDefaultsSchemaComponent

// #/components/schemas/WithConst
type WithConstSchemaComponent struct {
	Version *string `json:"version,omitempty"`
	_Type   *string `json:"type,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *WithConstSchemaComponent) ApplyDefaults() {
}

type WithConst = WithConstSchemaComponent

// #/components/schemas/WithConstraints
type WithConstraintsSchemaComponent struct {
	BoundedInt          *int     `json:"boundedInt,omitempty"`
	ExclusiveBoundedInt *int     `json:"exclusiveBoundedInt,omitempty"`
	MultipleOf          *int     `json:"multipleOf,omitempty"`
	BoundedString       *string  `json:"boundedString,omitempty"`
	PatternString       *string  `json:"patternString,omitempty"`
	BoundedArray        []string `json:"boundedArray,omitempty"`
	UniqueArray         []string `json:"uniqueArray,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *WithConstraintsSchemaComponent) ApplyDefaults() {
}

type WithConstraints = WithConstraintsSchemaComponent

// #/components/schemas/TypeArray31
type TypeArray31SchemaComponent struct {
	Name *string `json:"name,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *TypeArray31SchemaComponent) ApplyDefaults() {
}

type TypeArray31 = TypeArray31SchemaComponent

// #/components/schemas/ExplicitAny
type ExplicitAnySchemaComponent = string

type ExplicitAny = ExplicitAnySchemaComponent

// #/components/schemas/ComplexNested
type ComplexNestedSchemaComponent struct {
	Metadata map[string]any          `json:"metadata,omitempty"`
	Items    []ComplexNestedItemItem `json:"items,omitempty"`
	Config   *ComplexNestedConfig    `json:"config,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *ComplexNestedSchemaComponent) ApplyDefaults() {
}

type ComplexNested = ComplexNestedSchemaComponent

// #/components/schemas/ComplexNested/properties/metadata
type ComplexNestedMetadataPropertySchemaComponent = map[string]any

type ComplexNestedMetadata = ComplexNestedMetadataPropertySchemaComponent

// #/components/schemas/ComplexNested/properties/metadata/additionalProperties
type ComplexNestedMetadataValuePropertySchemaComponent struct {
	_String0        *string
	_Int1           *int
	LBracketString2 *[]string
}

func (u ComplexNestedMetadataValuePropertySchemaComponent) MarshalJSON() ([]byte, error) {
	if u._String0 != nil {
		return json.Marshal(u._String0)
	}
	if u._Int1 != nil {
		return json.Marshal(u._Int1)
	}
	if u.LBracketString2 != nil {
		return json.Marshal(u.LBracketString2)
	}
	return []byte("null"), nil
}

func (u *ComplexNestedMetadataValuePropertySchemaComponent) UnmarshalJSON(data []byte) error {
	var v0 string
	if err := json.Unmarshal(data, &v0); err == nil {
		u._String0 = &v0
	}

	var v1 int
	if err := json.Unmarshal(data, &v1); err == nil {
		u._Int1 = &v1
	}

	var v2 []string
	if err := json.Unmarshal(data, &v2); err == nil {
		u.LBracketString2 = &v2
	}

	return nil
}

// ApplyDefaults sets default values for fields that are nil.
func (u *ComplexNestedMetadataValuePropertySchemaComponent) ApplyDefaults() {
}

type ComplexNestedMetadataValue = ComplexNestedMetadataValuePropertySchemaComponent

// #/components/schemas/ComplexNested/properties/items
type ComplexNestedItemPropertySchemaComponent = []ComplexNestedItemItem

type ComplexNestedItem = ComplexNestedItemPropertySchemaComponent

// #/components/schemas/ComplexNested/properties/items/items
type ComplexNestedItemItemPropertySchemaComponent struct {
	ID        *int       `json:"id,omitempty"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	Tags      []string   `json:"tags,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *ComplexNestedItemItemPropertySchemaComponent) ApplyDefaults() {
}

type ComplexNestedItemItem = ComplexNestedItemItemPropertySchemaComponent

// #/components/schemas/ComplexNested/properties/items/items/allOf/1
type ComplexNestedN1AllOfItemItemPropertySchemaComponent struct {
	Tags []string `json:"tags,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *ComplexNestedN1AllOfItemItemPropertySchemaComponent) ApplyDefaults() {
}

type ComplexNestedAllOf1 = ComplexNestedN1AllOfItemItemPropertySchemaComponent

// #/components/schemas/ComplexNested/properties/config
type ComplexNestedConfigPropertySchemaComponent struct {
	ComplexNestedConfigOneOf0 *ComplexNestedConfigOneOf0
	ComplexNestedConfigOneOf1 *ComplexNestedConfigOneOf1
}

func (u ComplexNestedConfigPropertySchemaComponent) MarshalJSON() ([]byte, error) {
	var count int
	var data []byte
	var err error

	if u.ComplexNestedConfigOneOf0 != nil {
		count++
		data, err = json.Marshal(u.ComplexNestedConfigOneOf0)
		if err != nil {
			return nil, err
		}
	}
	if u.ComplexNestedConfigOneOf1 != nil {
		count++
		data, err = json.Marshal(u.ComplexNestedConfigOneOf1)
		if err != nil {
			return nil, err
		}
	}

	if count != 1 {
		return nil, fmt.Errorf("ComplexNestedConfigPropertySchemaComponent: exactly one member must be set, got %d", count)
	}

	return data, nil
}

func (u *ComplexNestedConfigPropertySchemaComponent) UnmarshalJSON(data []byte) error {
	var successCount int

	var v0 ComplexNestedConfigOneOf0
	if err := json.Unmarshal(data, &v0); err == nil {
		u.ComplexNestedConfigOneOf0 = &v0
		successCount++
	}

	var v1 ComplexNestedConfigOneOf1
	if err := json.Unmarshal(data, &v1); err == nil {
		u.ComplexNestedConfigOneOf1 = &v1
		successCount++
	}

	if successCount != 1 {
		return fmt.Errorf("ComplexNestedConfigPropertySchemaComponent: expected exactly one type to match, got %d", successCount)
	}

	return nil
}

// ApplyDefaults sets default values for fields that are nil.
func (u *ComplexNestedConfigPropertySchemaComponent) ApplyDefaults() {
	if u.ComplexNestedConfigOneOf0 != nil {
		u.ComplexNestedConfigOneOf0.ApplyDefaults()
	}
	if u.ComplexNestedConfigOneOf1 != nil {
		u.ComplexNestedConfigOneOf1.ApplyDefaults()
	}
}

type ComplexNestedConfig = ComplexNestedConfigPropertySchemaComponent

// #/components/schemas/ComplexNested/properties/config/oneOf/0
type ComplexNestedConfigN0OneOfPropertySchemaComponent struct {
	Mode  *string `json:"mode,omitempty"`
	Value *string `json:"value,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *ComplexNestedConfigN0OneOfPropertySchemaComponent) ApplyDefaults() {
}

type ComplexNestedConfigOneOf0 = ComplexNestedConfigN0OneOfPropertySchemaComponent

// #/components/schemas/ComplexNested/properties/config/oneOf/1
type ComplexNestedConfigN1OneOfPropertySchemaComponent struct {
	Mode    *string           `json:"mode,omitempty"`
	Options map[string]string `json:"options,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *ComplexNestedConfigN1OneOfPropertySchemaComponent) ApplyDefaults() {
}

type ComplexNestedConfigOneOf1 = ComplexNestedConfigN1OneOfPropertySchemaComponent

// #/components/schemas/ComplexNested/properties/config/oneOf/1/properties/options
type ComplexNestedConfigN1OptionsPropertyOneOfPropertySchemaComponent = map[string]string

type ComplexNestedConfigOneOf1Options = ComplexNestedConfigN1OptionsPropertyOneOfPropertySchemaComponent

// #/components/schemas/StringMap
type StringMapSchemaComponent = map[string]string

type StringMap = StringMapSchemaComponent

// #/components/schemas/ObjectMap
type ObjectMapSchemaComponent = map[string]any

type ObjectMap = ObjectMapSchemaComponent

// #/components/schemas/NestedMap
type NestedMapSchemaComponent = map[string]map[string]string

type NestedMap = NestedMapSchemaComponent

// #/components/schemas/NestedMap/additionalProperties
type NestedMapValueSchemaComponent = map[string]string

type NestedMapValue = NestedMapValueSchemaComponent

// #/paths//inline-response/get/responses/200/content/application/json/schema
type InlineResponseGetN200ApplicationJSONContentResponsePath struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *InlineResponseGetN200ApplicationJSONContentResponsePath) ApplyDefaults() {
}

type GetInlineResponse200Response = InlineResponseGetN200ApplicationJSONContentResponsePath

// #/paths//multi-content/post/requestBody/content/multipart/form-data/schema
type MultiContentPostMultipartFormDataContentRequestPath struct {
	File     *File   `json:"file,omitempty"`
	Metadata *string `json:"metadata,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *MultiContentPostMultipartFormDataContentRequestPath) ApplyDefaults() {
}

type PostMultiContentRequestForm2 = MultiContentPostMultipartFormDataContentRequestPath

const (
	emailRegexString = "^(?:(?:(?:(?:[a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(?:\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|(?:(?:\\x22)(?:(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(?:\\x20|\\x09)+)?(?:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(\\x20|\\x09)+)?(?:\\x22))))@(?:(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$"
)

var (
	emailRegex = regexp.MustCompile(emailRegexString)
)

// ErrValidationEmail is the sentinel error returned when an email fails validation
var ErrValidationEmail = errors.New("email: failed to pass regex validation")

// Email represents an email address.
// It is a string type that must pass regex validation before being marshalled
// to JSON or unmarshalled from JSON.
type Email string

func (e Email) MarshalJSON() ([]byte, error) {
	if !emailRegex.MatchString(string(e)) {
		return nil, ErrValidationEmail
	}

	return json.Marshal(string(e))
}

func (e *Email) UnmarshalJSON(data []byte) error {
	if e == nil {
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	*e = Email(s)
	if !emailRegex.MatchString(s) {
		return ErrValidationEmail
	}

	return nil
}

type File struct {
	multipart *multipart.FileHeader
	data      []byte
	filename  string
}

func (file *File) InitFromMultipart(header *multipart.FileHeader) {
	file.multipart = header
	file.data = nil
	file.filename = ""
}

func (file *File) InitFromBytes(data []byte, filename string) {
	file.data = data
	file.filename = filename
	file.multipart = nil
}

func (file File) MarshalJSON() ([]byte, error) {
	b, err := file.Bytes()
	if err != nil {
		return nil, err
	}
	return json.Marshal(b)
}

func (file *File) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &file.data)
}

func (file File) Bytes() ([]byte, error) {
	if file.multipart != nil {
		f, err := file.multipart.Open()
		if err != nil {
			return nil, err
		}
		defer func() { _ = f.Close() }()
		return io.ReadAll(f)
	}
	return file.data, nil
}

func (file File) Reader() (io.ReadCloser, error) {
	if file.multipart != nil {
		return file.multipart.Open()
	}
	return io.NopCloser(bytes.NewReader(file.data)), nil
}

func (file File) Filename() string {
	if file.multipart != nil {
		return file.multipart.Filename
	}
	return file.filename
}

func (file File) FileSize() int64 {
	if file.multipart != nil {
		return file.multipart.Size
	}
	return int64(len(file.data))
}

const DateFormat = "2006-01-02"

type Date struct {
	time.Time
}

func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format(DateFormat))
}

func (d *Date) UnmarshalJSON(data []byte) error {
	var dateStr string
	err := json.Unmarshal(data, &dateStr)
	if err != nil {
		return err
	}
	parsed, err := time.Parse(DateFormat, dateStr)
	if err != nil {
		return err
	}
	d.Time = parsed
	return nil
}

func (d Date) String() string {
	return d.Format(DateFormat)
}

func (d *Date) UnmarshalText(data []byte) error {
	parsed, err := time.Parse(DateFormat, string(data))
	if err != nil {
		return err
	}
	d.Time = parsed
	return nil
}

type UUID = uuid.UUID
