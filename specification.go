package specifications

// Specification is the interface that all specifications must implement.
// This interface represents a condition or a set of conditions that can be
// translated by a visitor or applied to in-memory objects.
type Specification interface {
	// Accept allows a visitor to process the specification.
	Accept(v SpecificationVisitor)
}

// SpecificationVisitor defines how a specification can be translated into something
// else, for example a database query. Each specification calls its corresponding
// visitor method with its internal data.
type SpecificationVisitor interface {
	VisitEqual(field string, value interface{})
	VisitNotEqual(field string, value interface{})
	VisitIn(field string, values []interface{})
	VisitAnd(specs []Specification)
	VisitOr(specs []Specification)
	VisitLimit(limit int)
	VisitOrder(field, direction string)
	VisitGreaterThan(field string, value interface{})
	VisitLowerThan(field string, value interface{})
	VisitLike(field string, value interface{})
	VisitGreaterThanOrEqual(field string, value interface{})
	VisitLowerThanOrEqual(field string, value interface{})
	VisitOffset(offset int)
}

// Base structure to define atomic specifications (e.g. equality checks)
type equalSpec struct {
	field string
	value interface{}
}

func (s *equalSpec) Accept(v SpecificationVisitor) {
	v.VisitEqual(s.field, s.value)
}

// Base structure for "in" specification (if you need it)
type inSpec struct {
	field  string
	values []interface{}
}

func (s *inSpec) Accept(v SpecificationVisitor) {
	v.VisitIn(s.field, s.values)
}

// Composite specifications
type andSpec struct {
	specs []Specification
}

func (s *andSpec) Accept(v SpecificationVisitor) {
	v.VisitAnd(s.specs)
}

type orSpec struct {
	specs []Specification
}

func (s *orSpec) Accept(v SpecificationVisitor) {
	v.VisitOr(s.specs)
}

type limitSpec struct {
	limit int
}

func (s *limitSpec) Accept(v SpecificationVisitor) {
	v.VisitLimit(s.limit)
}

type orderSpec struct {
	field     string
	direction string
}

func (s *orderSpec) Accept(v SpecificationVisitor) {
	v.VisitOrder(s.field, s.direction)
}

type greaterThanSpec struct {
	field string
	value interface{}
}

func (s *greaterThanSpec) Accept(v SpecificationVisitor) {
	v.VisitGreaterThan(s.field, s.value)
}

type lowerThanSpec struct {
	field string
	value interface{}
}

func (s *lowerThanSpec) Accept(v SpecificationVisitor) {
	v.VisitLowerThan(s.field, s.value)
}

type greaterThanOrEqualSpec struct {
	field string
	value interface{}
}

func (s *greaterThanOrEqualSpec) Accept(v SpecificationVisitor) {
	v.VisitGreaterThanOrEqual(s.field, s.value)
}

type lowerThanOrEqualSpec struct {
	field string
	value interface{}
}

func (s *lowerThanOrEqualSpec) Accept(v SpecificationVisitor) {
	v.VisitLowerThanOrEqual(s.field, s.value)
}

type notEqualSpec struct {
	field string
	value interface{}
}

func (s *notEqualSpec) Accept(v SpecificationVisitor) {
	v.VisitNotEqual(s.field, s.value)
}

type likeSpec struct {
	field string
	value interface{}
}

func (s *likeSpec) Accept(v SpecificationVisitor) {
	v.VisitLike(s.field, s.value)
}

type offsetSpec struct {
	offset int
}

func (s *offsetSpec) Accept(v SpecificationVisitor) {
	v.VisitOffset(s.offset)
}

func GreaterThanOrEqual(field string, value interface{}) Specification {
	return &greaterThanOrEqualSpec{
		field: field,
		value: value,
	}
}

func LowerThanOrEqual(field string, value interface{}) Specification {
	return &lowerThanOrEqualSpec{
		field: field,
		value: value,
	}
}

func NotEqual(field string, value interface{}) Specification {
	return &notEqualSpec{
		field: field,
		value: value,
	}
}

func GreaterThan(field string, value interface{}) Specification {
	return &greaterThanSpec{
		field: field,
		value: value,
	}
}

func LowerThan(field string, value interface{}) Specification {
	return &lowerThanSpec{
		field: field,
		value: value,
	}
}

func Like(field string, value interface{}) Specification {
	return &likeSpec{
		field: field,
		value: value,
	}
}

func Offset(offset int) Specification {
	return &offsetSpec{
		offset: offset,
	}
}

func Equal(field string, value interface{}) Specification {
	return &equalSpec{
		field: field,
		value: value,
	}
}

func In(field string, values ...interface{}) Specification {
	return &inSpec{
		field:  field,
		values: values,
	}
}

func And(specs ...Specification) Specification {
	return &andSpec{specs: specs}
}

func Or(specs ...Specification) Specification {
	return &orSpec{specs: specs}
}

func Limit(limit int) Specification {
	return &limitSpec{limit: limit}
}

func OrderBy(field string, direction string) Specification {
	return &orderSpec{
		field:     field,
		direction: direction,
	}
}
