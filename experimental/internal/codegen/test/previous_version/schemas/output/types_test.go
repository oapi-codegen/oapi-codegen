package output

import (
	"encoding/json"
	"testing"
)

func TestGenericObjectInstantiation(t *testing.T) {
	g := GenericObject{}
	_ = g
}

func TestAnyTypeAliases(t *testing.T) {
	var a1 AnyType1 = "hello"
	var a2 AnyType2 = 42
	var cs CustomStringType = "test"
	_, _, _ = a1, a2, cs
}

func TestNullablePropertiesInstantiation(t *testing.T) {
	opt := "optional"
	np := NullableProperties{
		Optional:            &opt,
		OptionalAndNullable: NewNullableWithValue("nullable"),
		Required:            "required",
		RequiredAndNullable: NewNullNullable[string](),
	}
	if np.Required != "required" {
		t.Errorf("unexpected required: %s", np.Required)
	}
	if np.Optional == nil || *np.Optional != "optional" {
		t.Errorf("unexpected optional: %v", np.Optional)
	}
	v, err := np.OptionalAndNullable.Get()
	if err != nil || v != "nullable" {
		t.Errorf("unexpected optionalAndNullable: %v, %v", v, err)
	}
	if !np.RequiredAndNullable.IsNull() {
		t.Error("requiredAndNullable should be null")
	}
}

func TestNullablePropertiesJSONRoundTrip(t *testing.T) {
	np := NullableProperties{
		Required:            "req",
		RequiredAndNullable: NewNullableWithValue("val"),
	}
	data, err := json.Marshal(np)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded NullableProperties
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Required != "req" {
		t.Errorf("round-trip failed for required: %s", decoded.Required)
	}
}

func TestN5StartsWithNumberInstantiation(t *testing.T) {
	n := N5StartsWithNumber{}
	_ = n
}

func TestEnumInObjInArrayType(t *testing.T) {
	item := EnumInObjInArrayItem{Val: strPtr("first")}
	arr := EnumInObjInArray{item}
	if len(arr) != 1 {
		t.Errorf("expected 1 item, got %d", len(arr))
	}
}

func TestEnumInObjInArrayConstants(t *testing.T) {
	if First != "first" {
		t.Errorf("unexpected first: %s", First)
	}
	if Second != "second" {
		t.Errorf("unexpected second: %s", Second)
	}
}

func TestDeprecatedPropertyInstantiation(t *testing.T) {
	dp := DeprecatedProperty{
		NewProp: "new",
	}
	if dp.NewProp != "new" {
		t.Errorf("unexpected newProp: %s", dp.NewProp)
	}
}

func TestOuterTypeWithAnonymousInner(t *testing.T) {
	o := OuterTypeWithAnonymousInner{
		Name:  "outer",
		Inner: InnerRenamedAnonymousObject{ID: 42},
	}
	if o.Name != "outer" {
		t.Errorf("unexpected name: %s", o.Name)
	}
	if o.Inner.ID != 42 {
		t.Errorf("unexpected inner id: %d", o.Inner.ID)
	}
}

func TestOuterTypeWithAnonymousInnerJSONRoundTrip(t *testing.T) {
	o := OuterTypeWithAnonymousInner{
		Name:  "test",
		Inner: InnerRenamedAnonymousObject{ID: 7},
	}
	data, err := json.Marshal(o)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded OuterTypeWithAnonymousInner
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Name != "test" || decoded.Inner.ID != 7 {
		t.Errorf("round-trip failed: %+v", decoded)
	}
}

func TestApplyDefaults(t *testing.T) {
	g := &GenericObject{}
	g.ApplyDefaults()
	np := &NullableProperties{}
	np.ApplyDefaults()
	n5 := &N5StartsWithNumber{}
	n5.ApplyDefaults()
	dp := &DeprecatedProperty{}
	dp.ApplyDefaults()
	o := &OuterTypeWithAnonymousInner{}
	o.ApplyDefaults()
}

func TestGetOpenAPISpecJSON(t *testing.T) {
	data, err := GetOpenAPISpecJSON()
	if err != nil {
		t.Fatalf("GetOpenAPISpecJSON failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetOpenAPISpecJSON returned empty data")
	}
}

func strPtr(s string) *string {
	return &s
}
