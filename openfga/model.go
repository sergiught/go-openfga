package openfga

// This file defines the strongly-typed authorization-model schema and small
// builder helpers for authoring models in Go. The JSON field names mirror
// OpenFGA's model schema exactly, including its mix of camelCase (computedUserset,
// tupleToUserset) and snake_case (directly_related_user_types). The types
// round-trip losslessly with the DSL transformer in the sibling dsl module, so
// you can author with either.

// Userset is a relation's rewrite rule. Exactly one field is set. Use the
// builder helpers (This, ComputedUserset, TupleToUserset, Union, Intersection,
// Difference) rather than populating the fields by hand.
type Userset struct {
	This            *DirectUserset  `json:"this,omitempty"`
	ComputedUserset *ObjectRelation `json:"computedUserset,omitempty"`
	TupleToUserset  *TupleToUserset `json:"tupleToUserset,omitempty"`
	Union           *Usersets       `json:"union,omitempty"`
	Intersection    *Usersets       `json:"intersection,omitempty"`
	Difference      *Difference     `json:"difference,omitempty"`
}

// DirectUserset marks a relation as directly assignable ("this" in the DSL, the
// `[...]` type restriction). It serializes as an empty JSON object.
type DirectUserset struct{}

// ObjectRelation references a relation, optionally on a specific object.
type ObjectRelation struct {
	Object   string `json:"object,omitempty"`
	Relation string `json:"relation,omitempty"`
}

// TupleToUserset rewrites through the Tupleset relation to the ComputedUserset
// relation on the resolved objects ("X from Y" in the DSL).
type TupleToUserset struct {
	Tupleset        ObjectRelation `json:"tupleset"`
	ComputedUserset ObjectRelation `json:"computedUserset"`
}

// Usersets holds the operands of a Union or Intersection.
type Usersets struct {
	Child []Userset `json:"child"`
}

// Difference is Base minus Subtract ("A but not B" in the DSL).
type Difference struct {
	Base     Userset `json:"base"`
	Subtract Userset `json:"subtract"`
}

// Metadata carries per-relation typing information for a type definition.
type Metadata struct {
	Relations  map[string]RelationMetadata `json:"relations,omitempty"`
	Module     string                      `json:"module,omitempty"`
	SourceInfo *SourceInfo                 `json:"source_info,omitempty"`
}

// RelationMetadata lists the user types that may be directly assigned to a
// relation (the `[...]` restriction in the DSL).
type RelationMetadata struct {
	DirectlyRelatedUserTypes []RelationReference `json:"directly_related_user_types,omitempty"`
	Module                   string              `json:"module,omitempty"`
	SourceInfo               *SourceInfo         `json:"source_info,omitempty"`
}

// RelationReference is one allowed user type for a relation: a bare type
// (`user`), a userset (`group#member`, via Relation), a type-bound wildcard
// (`user:*`, via Wildcard), or any of these gated by a condition (via Condition).
type RelationReference struct {
	Type      string    `json:"type"`
	Relation  string    `json:"relation,omitempty"`
	Wildcard  *Wildcard `json:"wildcard,omitempty"`
	Condition string    `json:"condition,omitempty"`
}

// Wildcard marks a type-bound wildcard reference (`user:*`). It serializes as an
// empty JSON object.
type Wildcard struct{}

// SourceInfo is DSL source-position metadata the transformer may attach. It is
// preserved on round-trips; you do not set it when authoring a model by hand.
type SourceInfo struct {
	File string `json:"file,omitempty"`
}

// Condition is an ABAC condition referenced by tuples and relation references.
type Condition struct {
	Name       string                        `json:"name"`
	Expression string                        `json:"expression"`
	Parameters map[string]ConditionParamType `json:"parameters,omitempty"`
	Metadata   *ConditionMetadata            `json:"metadata,omitempty"`
}

// ConditionParamType is the type of a condition parameter, e.g. TYPE_NAME_INT.
// GenericTypes carries the element type(s) for container types such as
// TYPE_NAME_LIST and TYPE_NAME_MAP.
type ConditionParamType struct {
	TypeName     string               `json:"type_name"`
	GenericTypes []ConditionParamType `json:"generic_types,omitempty"`
}

// ConditionMetadata is optional source metadata attached to a condition.
type ConditionMetadata struct {
	Module     string      `json:"module,omitempty"`
	SourceInfo *SourceInfo `json:"source_info,omitempty"`
}

// Builder helpers. Each returns a Userset with exactly one branch set, so
// relation rewrites read close to the DSL:
//
//	Relations: map[string]openfga.Userset{
//		"owner":  openfga.This(),
//		"editor": openfga.Union(openfga.This(), openfga.ComputedUserset("owner")),
//		"viewer": openfga.TupleTo("parent", "viewer"),
//	}

// This returns a directly-assignable rewrite (the DSL `[...]` restriction).
func This() Userset { return Userset{This: &DirectUserset{}} }

// ComputedUserset returns a rewrite to another relation on the same object
// (the DSL bare-relation reference, e.g. `owner`).
func ComputedUserset(relation string) Userset {
	return Userset{ComputedUserset: &ObjectRelation{Relation: relation}}
}

// TupleTo returns a rewrite that follows the tupleset relation and then applies
// computedRelation on the resolved objects (the DSL `computedRelation from
// tupleset`). It builds a TupleToUserset.
func TupleTo(tupleset, computedRelation string) Userset {
	return Userset{TupleToUserset: &TupleToUserset{
		Tupleset:        ObjectRelation{Relation: tupleset},
		ComputedUserset: ObjectRelation{Relation: computedRelation},
	}}
}

// Union returns the union of the given rewrites (the DSL `or`).
func Union(children ...Userset) Userset {
	return Userset{Union: &Usersets{Child: children}}
}

// Intersection returns the intersection of the given rewrites (the DSL `and`).
func Intersection(children ...Userset) Userset {
	return Userset{Intersection: &Usersets{Child: children}}
}

// Exclusion returns base minus subtract (the DSL `but not`). It builds a
// Difference.
func Exclusion(base, subtract Userset) Userset {
	return Userset{Difference: &Difference{Base: base, Subtract: subtract}}
}

// DirectType returns a directly-related user type for a relation's metadata,
// e.g. DirectType("user") for `[user]`. Set Relation for a userset
// (`group#member`), Wildcard for `user:*`, or Condition to gate the reference.
func DirectType(typ string) RelationReference { return RelationReference{Type: typ} }
