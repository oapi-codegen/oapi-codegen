package aggregatesoneof

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue1530(t *testing.T) {
	httpConfigTypes := []string{
		"another_server",
		"apache_server",
		"web_server",
	}

	for _, configType := range httpConfigTypes {
		t.Run("http-"+configType, func(t *testing.T) {
			saveReq := ConfigSaveReq{}
			err := saveReq.FromConfigHttp(ConfigHttp{
				ConfigType: configType,
				Host:       "example.com",
			})
			require.NoError(t, err)

			cfg, err := saveReq.AsConfigHttp()
			require.NoError(t, err)
			require.Equal(t, configType, cfg.ConfigType)

			cfgByDiscriminator, err := saveReq.ValueByDiscriminator()
			require.NoError(t, err)
			require.Equal(t, cfg, cfgByDiscriminator)
		})
	}

	t.Run("ssh", func(t *testing.T) {
		saveReq := ConfigSaveReq{}
		err := saveReq.FromConfigSsh(ConfigSsh{
			ConfigType: "ssh_server",
		})
		require.NoError(t, err)

		cfg, err := saveReq.AsConfigSsh()
		require.NoError(t, err)
		require.Equal(t, "ssh_server", cfg.ConfigType)

		cfgByDiscriminator, err := saveReq.ValueByDiscriminator()
		require.NoError(t, err)
		require.Equal(t, cfg, cfgByDiscriminator)
	})
}

// TestIssue2297PointerDiscriminatorOnVariant covers a discriminator property
// that is optional on the variants and narrowed to a single-value enum, so it
// renders as a pointer to a named enum type. From*/Merge* stamp the
// discriminator into the union JSON without touching the variant's field, so
// the caller doesn't need to populate it.
func TestIssue2297PointerDiscriminatorOnVariant(t *testing.T) {
	var conflict ConflictError
	require.NoError(t, conflict.FromResourceConflictError(ResourceConflictError{Error: "already there"}))

	d, err := conflict.Discriminator()
	require.NoError(t, err)
	require.Equal(t, "resource_exists", d)

	v, err := conflict.ValueByDiscriminator()
	require.NoError(t, err)
	rce, ok := v.(ResourceConflictError)
	require.True(t, ok)
	require.NotNil(t, rce.Code)
	require.Equal(t, ResourceExists, *rce.Code)
	require.Equal(t, "already there", rce.Error)

	// Merging the other variant re-stamps the discriminator.
	require.NoError(t, conflict.MergeIdempotencyConflictError(IdempotencyConflictError{Error: "replayed"}))
	d, err = conflict.Discriminator()
	require.NoError(t, err)
	require.Equal(t, "idempotency_conflict", d)
}

// TestIssue2297PointerDiscriminatorOnUnion covers the union schema declaring
// the discriminator property itself, optional, so the field on the union
// struct is pointer-typed and is set through an addressable value. The value
// is also stamped into the union data, so Discriminator() and
// ValueByDiscriminator() work immediately after From*.
func TestIssue2297PointerDiscriminatorOnUnion(t *testing.T) {
	var pet PetByKind
	meow := "prrr"
	require.NoError(t, pet.FromKindCat(KindCat{Meow: &meow}))

	require.NotNil(t, pet.Kind)
	require.Equal(t, "cat", *pet.Kind)

	d, err := pet.Discriminator()
	require.NoError(t, err)
	require.Equal(t, "cat", d)

	v, err := pet.ValueByDiscriminator()
	require.NoError(t, err)
	cat, ok := v.(KindCat)
	require.True(t, ok)
	require.NotNil(t, cat.Meow)
	require.Equal(t, meow, *cat.Meow)

	b, err := pet.MarshalJSON()
	require.NoError(t, err)
	require.Contains(t, string(b), `"kind":"cat"`)
}

// TestIssue2297RenamedDiscriminatorField covers a union whose discriminator
// property is renamed via x-go-name: the field is matched by its JSON name,
// assigned through the renamed Go field, and stamped into the union data.
func TestIssue2297RenamedDiscriminatorField(t *testing.T) {
	var pet RenamedPetByKind
	bark := "woof"
	require.NoError(t, pet.FromKindDog(KindDog{Bark: &bark}))

	require.NotNil(t, pet.Species)
	require.Equal(t, "dog", *pet.Species)

	d, err := pet.Discriminator()
	require.NoError(t, err)
	require.Equal(t, "dog", d)

	b, err := pet.MarshalJSON()
	require.NoError(t, err)
	require.Contains(t, string(b), `"kind":"dog"`)
}
