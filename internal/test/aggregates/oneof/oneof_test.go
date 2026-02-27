package oneof

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicOneOfWithObject(t *testing.T) {
	v := Variant1{Name: "test"}

	var o BasicOneOf
	err := o.FromVariant1(v)
	require.NoError(t, err)

	data, err := json.Marshal(o)
	require.NoError(t, err)

	var roundTripped BasicOneOf
	err = json.Unmarshal(data, &roundTripped)
	require.NoError(t, err)

	got, err := roundTripped.AsVariant1()
	require.NoError(t, err)
	assert.Equal(t, "test", got.Name)
}

func TestBasicOneOfWithArray(t *testing.T) {
	var o BasicOneOf
	err := o.FromVariant2(Variant2{1, 2, 3})
	require.NoError(t, err)

	data, err := json.Marshal(o)
	require.NoError(t, err)

	var roundTripped BasicOneOf
	err = json.Unmarshal(data, &roundTripped)
	require.NoError(t, err)

	got, err := roundTripped.AsVariant2()
	require.NoError(t, err)
	assert.Equal(t, Variant2{1, 2, 3}, got)
}

func TestBasicOneOfWithBoolean(t *testing.T) {
	var o BasicOneOf
	err := o.FromVariant3(true)
	require.NoError(t, err)

	data, err := json.Marshal(o)
	require.NoError(t, err)
	assert.Equal(t, `true`, string(data))
}

func TestInlineOneOfWithObject(t *testing.T) {
	label := "hello"
	var o InlineOneOf
	err := o.FromInlineOneOf0(InlineOneOf0{Label: &label})
	require.NoError(t, err)

	data, err := json.Marshal(o)
	require.NoError(t, err)

	var roundTripped InlineOneOf
	err = json.Unmarshal(data, &roundTripped)
	require.NoError(t, err)

	got, err := roundTripped.AsInlineOneOf0()
	require.NoError(t, err)
	assert.Equal(t, "hello", *got.Label)
}

func TestOneOfWithDiscriminator(t *testing.T) {
	va := DiscriminatedVariantA{Discriminator: "DiscriminatedVariantA", Name: "test-a"}
	var o OneOfWithDiscriminator
	err := o.FromDiscriminatedVariantA(va)
	require.NoError(t, err)

	disc, err := o.Discriminator()
	require.NoError(t, err)
	assert.Equal(t, "DiscriminatedVariantA", disc)

	val, err := o.ValueByDiscriminator()
	require.NoError(t, err)
	got, ok := val.(DiscriminatedVariantA)
	require.True(t, ok)
	assert.Equal(t, "test-a", got.Name)
}

func TestOneOfWithDiscriminatorVariantB(t *testing.T) {
	vb := DiscriminatedVariantB{Discriminator: "DiscriminatedVariantB", Id: 42}
	var o OneOfWithDiscriminator
	err := o.FromDiscriminatedVariantB(vb)
	require.NoError(t, err)

	disc, err := o.Discriminator()
	require.NoError(t, err)
	assert.Equal(t, "DiscriminatedVariantB", disc)

	val, err := o.ValueByDiscriminator()
	require.NoError(t, err)
	got, ok := val.(DiscriminatedVariantB)
	require.True(t, ok)
	assert.Equal(t, 42, got.Id)
}

func TestOneOfWithMapping(t *testing.T) {
	va := DiscriminatedVariantA{Discriminator: "va", Name: "mapped-a"}
	var o OneOfWithMapping
	err := o.FromDiscriminatedVariantA(va)
	require.NoError(t, err)

	disc, err := o.Discriminator()
	require.NoError(t, err)
	assert.Equal(t, "va", disc)

	val, err := o.ValueByDiscriminator()
	require.NoError(t, err)
	got, ok := val.(DiscriminatedVariantA)
	require.True(t, ok)
	assert.Equal(t, "mapped-a", got.Name)
}

func TestOneOfWithFixedProps(t *testing.T) {
	var o OneOfWithFixedProps
	err := o.FromVariant1(Variant1{Name: "test"})
	require.NoError(t, err)

	data, err := json.Marshal(o)
	require.NoError(t, err)

	var roundTripped OneOfWithFixedProps
	err = json.Unmarshal(data, &roundTripped)
	require.NoError(t, err)

	got, err := roundTripped.AsVariant1()
	require.NoError(t, err)
	assert.Equal(t, "test", got.Name)
}

func TestOneOfWithFixedDiscriminator(t *testing.T) {
	v := Variant1{Name: "named"}
	var o OneOfWithFixedDiscriminator
	err := o.FromVariant1(v)
	require.NoError(t, err)

	data, err := json.Marshal(o)
	require.NoError(t, err)

	var roundTripped OneOfWithFixedDiscriminator
	err = json.Unmarshal(data, &roundTripped)
	require.NoError(t, err)

	got, err := roundTripped.AsVariant1()
	require.NoError(t, err)
	assert.Equal(t, "named", got.Name)
}

func TestArrayOfOneOf(t *testing.T) {
	var item1 ArrayOfOneOf_Item
	err := item1.FromVariant1(Variant1{Name: "first"})
	require.NoError(t, err)

	var item2 ArrayOfOneOf_Item
	err = item2.FromVariant2(Variant2{10, 20})
	require.NoError(t, err)

	arr := ArrayOfOneOf{item1, item2}
	data, err := json.Marshal(arr)
	require.NoError(t, err)

	var roundTripped ArrayOfOneOf
	err = json.Unmarshal(data, &roundTripped)
	require.NoError(t, err)
	require.Len(t, roundTripped, 2)

	got1, err := roundTripped[0].AsVariant1()
	require.NoError(t, err)
	assert.Equal(t, "first", got1.Name)

	got2, err := roundTripped[1].AsVariant2()
	require.NoError(t, err)
	assert.Equal(t, Variant2{10, 20}, got2)
}

// From issue-1530: multiple discriminator values mapping to the same type.
func TestOneOfMultiMappingHttpVariants(t *testing.T) {
	testCases := []string{"apache_server", "web_server"}
	for _, configType := range testCases {
		t.Run(configType, func(t *testing.T) {
			http := ConfigHttp{ConfigType: configType, Url: "http://example.com"}
			var o OneOfMultiMapping
			err := o.FromConfigHttp(http)
			require.NoError(t, err)

			disc, err := o.Discriminator()
			require.NoError(t, err)
			assert.Equal(t, configType, disc)

			val, err := o.ValueByDiscriminator()
			require.NoError(t, err)
			gotHttp, ok := val.(ConfigHttp)
			require.True(t, ok)
			assert.Equal(t, "http://example.com", gotHttp.Url)
		})
	}
}

func TestOneOfMultiMappingSsh(t *testing.T) {
	ssh := ConfigSsh{ConfigType: "ssh_server", Host: "server.example.com"}
	var o OneOfMultiMapping
	err := o.FromConfigSsh(ssh)
	require.NoError(t, err)

	disc, err := o.Discriminator()
	require.NoError(t, err)
	assert.Equal(t, "ssh_server", disc)

	val, err := o.ValueByDiscriminator()
	require.NoError(t, err)
	gotSsh, ok := val.(ConfigSsh)
	require.True(t, ok)
	assert.Equal(t, "server.example.com", gotSsh.Host)
}

func TestOneOfMarshalWithNoValueSet(t *testing.T) {
	var o BasicOneOf
	data, err := json.Marshal(o)
	require.NoError(t, err)
	assert.Equal(t, "null", string(data))
}
