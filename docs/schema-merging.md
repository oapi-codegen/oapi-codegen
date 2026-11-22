# How `allOf`, `anyOf` and `oneOf` become Go types

OpenAPI composes schemas with `allOf` (a value must match every schema listed), `anyOf` (at least one) and `oneOf` (exactly one). Go has no composition, so `oapi-codegen` has to compute an appropriate Go model for each. The `compatibility.schema-merging-behavior` setting selects how it does that:

| Version | Status | `allOf` | `anyOf` and `oneOf` |
|---|---|---|---|
| `v1` | The original behavior, the only one before v1.11.0. Very limited. | Embeds each `$ref` member's type in a struct, and copies inline members' fields into it. | As in `v2`. |
| `v2` | The default since v1.11.0. | Merges the members into one Go type, keyword by keyword. | A union type with `As*`, `From*` and `Merge*` methods for each variant. |
| `v3` | Added in v2.9.0 as an option. It becomes the default once it has been tested for a while, likely in v2.10.0. | Merges the members by what `allOf` means, and reports conflicts as errors. | Union types that model more of what the spec says. |

To choose one, set it in your configuration:

```yaml
compatibility:
  schema-merging-behavior: v3
```

Leaving it unset means `v2`. `old-merge-schemas: true` is a deprecated alias for `v1`. `old-allof-sibling-merging` applies only to `v1` and `v2`; setting it with `v3` is a configuration error.

`v1` and `v2` still receive bug fixes, but not changes that would break code built on what they generate for a spec that already worked.

- [`v1`: embedding](#v1-embedding)
- [`v2` and `v3`: composition](#v2-and-v3-composition)
  - [`allOf`](#allof)
  - [`anyOf` and `oneOf`](#anyof-and-oneof)
- [Where `v3` reads the spec generously](#where-v3-reads-the-spec-generously)
- [Moving from `v2` to `v3`](#moving-from-v2-to-v3)

## `v1`: embedding

`v1` builds an `allOf` out of the Go types of its members: each `$ref` member's type is embedded in a struct, and the fields of inline members are added next to it. `v2` and `v3` instead compose the members' schemas into one schema, and generate one type from it.

```yaml
components:
  schemas:
    Client:
      type: object
      required: [name]
      properties:
        name: {type: string}
    Identity:
      type: object
      required: [issuer]
      properties:
        issuer: {type: string}
    ClientWithId:
      allOf:
        - $ref: '#/components/schemas/Client'
        - properties:
            id: {type: integer}
          required: [id]
    IdentityWithDuplicateField:
      allOf:
        - $ref: '#/components/schemas/Identity'
        - properties:
            issuer:
              type: string
              description: The URL of the issuer.
              maxLength: 255
```

<table>
<tr><th><code>v1</code></th><th><code>v2</code> and <code>v3</code></th></tr>
<tr>
<td>

```go
type ClientWithId struct {
	// Embedded struct due to allOf(#/components/schemas/Client)
	Client `yaml:",inline"`
	// Embedded fields due to inline allOf schema
	Id int `json:"id"`
}

type IdentityWithDuplicateField struct {
	// Embedded struct due to allOf(#/components/schemas/Identity)
	Identity `yaml:",inline"`
	// Embedded fields due to inline allOf schema
	// Issuer The URL of the issuer.
	Issuer *string `json:"issuer,omitempty"`
}
```

</td>
<td>

```go
type ClientWithId struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type IdentityWithDuplicateField struct {
	// Issuer The URL of the issuer.
	Issuer string `json:"issuer"`
}
```

</td>
</tr>
</table>

Embedding keeps each member's type, so a `ClientWithId` contains a `Client`. But the members never meet:

- A property that several members declare becomes several fields. `IdentityWithDuplicateField` has an `Issuer` of its own and one inside `Identity`. The outer one hides the inner one from `encoding/json`, so `Identity.Issuer` is never read or written.
- A keyword in one member doesn't reach another member's property. Above, the second member declares `issuer` without `required`, so its field is an optional pointer, although `Identity` requires it.

`v1` generates `anyOf` and `oneOf` as `v2` does. Use it only for code that depends on the embedded types.

## `v2` and `v3`: composition

Both merge an `allOf` into one Go type, and turn `anyOf` and `oneOf` into union types. A union type holds the JSON value, and has, for each variant, `As*` to read the value as that variant, `From*` to set it, and `Merge*` to merge a variant into it. A discriminated union also has `Discriminator()` and `ValueByDiscriminator()`.

`v2`'s rules grew one issue at a time: it merges the members two at a time, and when they disagree, the last one usually wins. `v3` merges them by what `allOf` means, that a value matches every member, and reports a disagreement no value can satisfy as an error that names both members. It also reads a number of common idioms the way their authors meant them.

### `allOf`

| Keyword in several members | `v2` | `v3` |
|---|---|---|
| `properties` | the last member's declaration of a property replaces the others | merged, [see below](#a-property-declared-in-several-members) |
| `type` | an error unless the same | the types they have in common: `integer` narrows `number`. No common type is an error |
| `enum` | all the values | the values they have in common, [see below](#enums) |
| `const` | ignored | must be the same, and allowed by the `enum` |
| `required` | all of them | all of them |
| `items`, `additionalProperties` | `items` are merged; two different `additionalProperties` schemas are an error | merged like properties |
| `discriminator` | two are an error | the same one may repeat |
| a nested `allOf` member's own keywords | partly dropped | kept |

Descriptions, examples, validation constraints and nullability never conflict. A property, or the composition, is nullable when any member makes it so.

#### A property declared in several members

A member can refine a property another member declares. This is common for patch payloads:

```yaml
Base:
  type: object
  required: [name]
  properties:
    name: {type: string}
Patch:
  allOf:
    - $ref: '#/components/schemas/Base'
    - properties:
        name: {nullable: true}
```

<table>
<tr><th><code>v2</code></th><th><code>v3</code></th></tr>
<tr>
<td>

```go
type Patch struct {
	Name any `json:"name"`
}
```

The second member's `{nullable: true}` replaces `Base`'s `name`.

</td>
<td>

```go
type Patch struct {
	Name *string `json:"name"`
}
```

`Base`'s string, made nullable by the second member.

</td>
</tr>
</table>

In `v3`, the property is merged as an `allOf` of its declarations, so nested objects merge their properties the same way, and the declarations' descriptions and field-level extensions such as `x-go-name` are kept.

#### Conflicts are errors

When members declare a property with types no value can have, `v2` keeps the last one. `v3` reports an error:

```yaml
Base:
  type: object
  properties:
    age: {type: integer}
Person:
  allOf:
    - $ref: '#/components/schemas/Base'
    - properties:
        age: {type: string}
```

<table>
<tr><th><code>v2</code></th><th><code>v3</code></th></tr>
<tr>
<td>

```go
type Person struct {
	Age *string `json:"age,omitempty"`
}
```

</td>
<td>

```
allOf can't merge #/components/schemas/Base/properties/age (type integer)
with allOf/1/properties/age (type string): no value has both types
```

</td>
</tr>
</table>

Types that do overlap narrow instead: `allOf: [{type: number}, {type: integer}]` is an error in `v2` (`can not merge incompatible types: [number], [integer]`), and an `int` in `v3`.

#### Enums

`v2` adds up the values of the members' enums. A value has to match every member, so `v3` keeps the values they have in common:

```yaml
Color:
  type: string
  enum: [red, green, blue]
Cool:
  allOf:
    - $ref: '#/components/schemas/Color'
    - enum: [green, blue, violet]
```

<table>
<tr><th><code>v2</code></th><th><code>v3</code></th></tr>
<tr>
<td>

```go
const (
	CoolBlue   Cool = "blue"
	CoolGreen  Cool = "green"
	CoolRed    Cool = "red"
	CoolViolet Cool = "violet"
)
```

</td>
<td>

```go
const (
	CoolBlue  Cool = "blue"
	CoolGreen Cool = "green"
)
```

</td>
</tr>
</table>

Enums with no value in common are an error in `v3`. Some specs use `allOf` on purpose to add values to an enum. The `x-oapi-codegen-enum-merge: union` extension asks `v3` for that:

```yaml
Status:
  type: string
  enum: [active, inactive]
ExtendedStatus:
  x-oapi-codegen-enum-merge: union
  allOf:
    - $ref: '#/components/schemas/Status'
    - enum: [archived]
```

This generates `active`, `inactive` and `archived` in both versions; `v2` ignores the extension, since it always adds up the values. The extension goes on the schema that has the `allOf`, or on a member's property that refines a property. Validators still read the `allOf` literally, so a service that validates requests against the spec has to accept the extra values some other way.

#### An `allOf` that only annotates a `$ref`

OpenAPI 3.0 allows nothing next to a `$ref`, so specs wrap one in an `allOf` to add a description, or to make it nullable. `v2` merges such an `allOf` like any other, which makes a copy of the referenced type. `v3` generates the referenced type itself:

```yaml
Named:
  type: object
  required: [name]
  properties:
    name: {type: string}
Described:
  allOf:
    - $ref: '#/components/schemas/Named'
    - description: A Named with a description of its own.
Holder:
  type: object
  properties:
    owner:
      allOf:
        - $ref: '#/components/schemas/Named'
        - nullable: true
```

<table>
<tr><th><code>v2</code></th><th><code>v3</code></th></tr>
<tr>
<td>

```go
type Described struct {
	Name string `json:"name"`
}

type Holder struct {
	Owner *struct {
		Name string `json:"name"`
	} `json:"owner"`
}
```

</td>
<td>

```go
// Described A Named with a description of its own.
type Described = Named

type Holder struct {
	Owner *Named `json:"owner"`
}
```

</td>
</tr>
</table>

A member only annotates when it has no keyword that shapes a Go type: documentation, `nullable` or a 3.1 `"null"` type, `readOnly`, `writeOnly`, `default`, validation constraints, and extensions other than `x-go-type` and the enum-naming `x-enum-varnames` and `x-enumNames`, which make a new type. A member that restates the referenced schema's type, like `{type: object, description: …}`, counts too. `x-go-type-skip-optional-pointer` in an annotating member applies to the result, and `x-go-type-name` names an alias. A member with `properties`, `required`, an `enum` or a different `type` makes a new type, merged as above.

#### Schemas from other documents, and `x-go-type`

`v3` doesn't read the schema of an `allOf` member that is in another document, or that `x-go-type` replaces with a Go type. Another generator run owns the external schema's type, and an `x-go-type`'s fields are unknown, so neither can be merged into a new struct:

- Such a member plus members that only annotate it is that member's type. A property `allOf: [$ref other.yaml#/User, {nullable: true}]` is `*externalRef0.User`, and a component with it is an alias, `= externalRef0.User`.
- Such a member plus any member that shapes a Go type is an error. For a schema from another document, define the schema in this one instead; for `x-go-type`, give the composition an `x-go-type` of its own.

```yaml
Timestamp:
  type: string
  x-go-type: time.Time
Created:
  allOf:
    - $ref: '#/components/schemas/Timestamp'
    - description: When it was created.
```

<table>
<tr><th><code>v2</code></th><th><code>v3</code></th></tr>
<tr>
<td>

```go
type Created = string
```

`Timestamp`'s `x-go-type` is lost ([#1622](https://github.com/oapi-codegen/oapi-codegen/issues/1622)).

</td>
<td>

```go
// Created When it was created.
type Created = Timestamp
```

</td>
</tr>
</table>

External references inside a member, in its properties, items or `additionalProperties`, are just type names, and work in both versions.

### `anyOf` and `oneOf`

#### A list that only adds constraints

A `oneOf` or `anyOf` often says which properties must be present, rather than listing alternative types:

```yaml
Contact:
  type: object
  properties:
    email: {type: string}
    phone: {type: string}
  oneOf:
    - required: [email]
    - required: [phone]
```

<table>
<tr><th><code>v2</code></th><th><code>v3</code></th></tr>
<tr>
<td>

```go
type Contact struct {
	Email *string `json:"email,omitempty"`
	Phone *string `json:"phone,omitempty"`
	union json.RawMessage
}

type Contact0 = any
type Contact1 = any

func (t Contact) AsContact0() (Contact0, error)
func (t *Contact) FromContact0(v Contact0) error
func (t *Contact) MergeContact0(v Contact0) error
func (t Contact) AsContact1() (Contact1, error)
// …
```

</td>
<td>

```go
type Contact struct {
	Email *string `json:"email,omitempty"`
	Phone *string `json:"phone,omitempty"`
}
```

</td>
</tr>
</table>

`v3` generates no union for a list whose branches declare nothing that shapes a Go type besides `required` and a boolean `additionalProperties`, or only restate the schema's own type. A validator still enforces the constraint; the Go type just doesn't pretend it's a choice of types. A branch with a type of its own, a `$ref` into another document, `x-go-type` or `x-go-type-name` is a type, and the union stays. A local `$ref` branch is judged by the schema it refers to.

#### A child that is an `allOf` of its parent

Some generators write inheritance with the parent listing its children in a `oneOf`, and each child an `allOf` of the parent:

```yaml
Pet:
  type: object
  required: [petType]
  properties:
    petType: {type: string}
  discriminator:
    propertyName: petType
  oneOf:
    - $ref: '#/components/schemas/Cat'
    - $ref: '#/components/schemas/Dog'
Cat:
  allOf:
    - $ref: '#/components/schemas/Pet'
    - properties:
        meow: {type: string}
Dog:
  allOf:
    - $ref: '#/components/schemas/Pet'
    - properties:
        bark: {type: string}
```

<table>
<tr><th><code>v2</code></th><th><code>v3</code></th></tr>
<tr>
<td>

Merging `Pet` into `Cat` copies `Pet`'s `oneOf`, so every child is a union of all the children:

```go
func (t Cat) AsCat() (Cat, error)
func (t Cat) AsDog() (Dog, error)
func (t *Cat) FromDog(v Dog) error
func (t Cat) ValueByDiscriminator() (any, error)
// …, and the same for Dog
```

</td>
<td>

A `Cat` is already one of `Pet`'s branches, so it is a plain struct:

```go
type Cat struct {
	Meow    *string `json:"meow,omitempty"`
	PetType string  `json:"petType"`
}
```

</td>
</tr>
</table>

`Pet` is the union of its children in both versions.

#### An `allOf` of several unions

```yaml
Payment:
  oneOf:
    - $ref: '#/components/schemas/Card'
    - $ref: '#/components/schemas/Transfer'
Delivery:
  oneOf:
    - $ref: '#/components/schemas/Courier'
    - $ref: '#/components/schemas/Pickup'
Order:
  allOf:
    - $ref: '#/components/schemas/Payment'
    - $ref: '#/components/schemas/Delivery'
```

An `Order` is a payment and a delivery at once. `v2` makes one union of all four variants, in which every `From*` replaces the whole value. `v3` keeps each union: a variant's `From*` replaces only its own union's data.

<table>
<tr><th><code>v2</code></th><th><code>v3</code></th></tr>
<tr>
<td>

```go
var o Order
o.FromCard(Card{Card: "4111"})
o.FromCourier(Courier{Address: "1 Main St"})
// {"address": "1 Main St"}
// the card is gone
```

</td>
<td>

```go
var o Order
o.FromCard(Card{Card: "4111"})
o.FromCourier(Courier{Address: "1 Main St"})
// {"address": "1 Main St", "card": "4111"}
o.FromTransfer(Transfer{Iban: "DE00"})
// {"address": "1 Main St", "iban": "DE00"}
```

</td>
</tr>
</table>

A union that is a `$ref` also gets `As*` and `From*` of its own in `v3`: `o.AsPayment()` returns a `Payment` without the delivery's data, and `o.FromPayment(p)` sets the payment and keeps the delivery. A key that variants of several unions declare, such as a shared `id`, is overwritten but never removed. A discriminator belongs to the union whose variants it tells apart.

#### Discriminated inline variants

A discriminator maps its values to schema names, explicitly with `mapping` or by each `$ref`'s name. An inline variant has no name, so `v2` can't map it. `v3` maps it to the value it pins for the discriminator property, with a `const` or an `enum` of one value:

```yaml
Pet:
  oneOf:
    - type: object
      properties:
        petType: {type: string, enum: [cat]}
        meow: {type: string}
    - type: object
      properties:
        petType: {type: string, enum: [dog]}
        bark: {type: string}
  discriminator:
    propertyName: petType
```

<table>
<tr><th><code>v2</code></th><th><code>v3</code></th></tr>
<tr>
<td>

```
discriminator: not all schemas were mapped
```

</td>
<td>

`ValueByDiscriminator()` returns a `Pet0` for `"cat"` and a `Pet1` for `"dog"`, and `FromPet0` writes `"petType": "cat"`.

</td>
</tr>
</table>

An inline variant that pins no value has no discriminator value: `ValueByDiscriminator()` doesn't lead to it, and its `From*` writes none, but its `As*` and `From*` work. `v3` also:

- reads an `integer` or `number` discriminator as a number, so `1`, `1.0` and `1e0` of a `number` discriminator lead to the same variant. `v2` compares the JSON text.
- writes the discriminator value in `From*` for each variant exactly one value leads to. `v2` writes none at all when any variant has several values ([#2071](https://github.com/oapi-codegen/oapi-codegen/issues/2071)).

#### Variants with `additionalProperties: false`

`As*` decodes the union's value into the variant, and ignores the keys the variant doesn't declare, so it can't tell variants apart ([#668](https://github.com/oapi-codegen/oapi-codegen/issues/668)):

```yaml
Subject:
  oneOf:
    - $ref: '#/components/schemas/Base'
    - $ref: '#/components/schemas/Extended'
Base:
  type: object
  additionalProperties: false
  required: [id]
  properties:
    id: {type: string}
Extended:
  allOf:
    - $ref: '#/components/schemas/Base'
    - type: object
      required: [extra]
      properties:
        extra: {type: string}
```

<table>
<tr><th><code>v2</code></th><th><code>v3</code></th></tr>
<tr>
<td>

`AsBase()` of `{"id": "2", "extra": "more"}` returns a `Base`, and drops `extra`.

</td>
<td>

`AsBase()` of `{"id": "2", "extra": "more"}` returns the error `Base doesn't allow the property "extra"`.

</td>
</tr>
</table>

In `v3`, the `As*` of a variant with `additionalProperties: false` checks the value's top-level keys first. It allows the keys the variant declares, and those anything else in the object declares: the union's own properties, its discriminator, and, in an `allOf` of unions, the other unions' variants. Nested objects are decoded as before. A variant whose keys can't all be known, because it is from another document, has `x-go-type`, or has a union of its own, is read as before. So is every variant of a union in an `allOf` with another union that has such a variant, since that variant might declare any key.

To turn the check off, set:

```yaml
output-options:
  lenient-union-accessors: true
```

Older clients may want this, so that they don't fail when a newer server adds a property to a variant. `v1` and `v2` ignore the option.

## Where `v3` reads the spec generously

Some idioms are common in real specs, but read literally, they don't do what their authors meant, or reject every value. `v3` generates what the author meant. OpenAPI validators, such as kin-openapi's `openapi3filter`, read them literally, so if you validate against your spec, use the spelling in the last column, which `v2` and `v3` read the same way.

| Idiom | Literally | `v3` generates | Spelling validators also accept |
|---|---|---|---|
| 3.0: `allOf: [$ref A, {nullable: true}]` | `null` must match `A` too, which rejects it | `*A` | `nullable: true` next to the `allOf`: `{nullable: true, allOf: [$ref A]}` |
| 3.1: `allOf: [$ref A, {type: 'null'}]`, or a member whose `type` includes `"null"` | `null` must match `A` too, and with `type: 'null'` nothing else matches, so no value does | `*A` | `anyOf: [$ref A, {type: 'null'}]` |
| `additionalProperties: false` in one `allOf` member | the member rejects the other members' properties, so no value with them matches | the whole object is closed | declare all the properties in the closed schema |
| a `oneOf` variant with `additionalProperties: false` | the variant rejects the union's own properties and discriminator | `As*` allows them | declare those properties in the variant too |
| `x-oapi-codegen-enum-merge: union` | the enums intersect | all the values | none: allow the extra values in your service |

## Moving from `v2` to `v3`

Set `schema-merging-behavior: v3`, regenerate, and look at what changed:

- **Errors.** `v3` reports specs whose `allOf` members contradict each other, which `v2` resolved by keeping the last member. Fix the spec. For enums, add `x-oapi-codegen-enum-merge: union` if the `allOf` is meant to add values. An external or `x-go-type` member combined with properties needs the schema defined in this document, or an `x-go-type` for the composition.
- **Aliases.** A component that only annotates a `$ref` is now an alias, `type Described = Named`. Conversions between them are no longer needed, and methods you declared on it now belong to the referenced type.
- **Fewer unions.** Constraint-only `oneOf`/`anyOf` lists and children of a parent that lists them are plain structs. Code that called `As*`/`From*` on them uses the fields instead.
- **Fewer enum values.** Enums combined with `allOf` keep only their common values.
- **Stricter `As*`.** Variants with `additionalProperties: false` reject other variants' data. Set `output-options.lenient-union-accessors` to keep `v2`'s behavior.
- **`allOf` of unions.** `From*` keeps the other unions' data. Inline variants of one of several unions are named `<Name>OneOf<n><index>` or `<Name>AnyOf<n><index>`, where `<n>` is only there when there are several `oneOf`s or `anyOf`s.
