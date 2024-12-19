package postgres

import (
	"fmt"
	"strings"

	"github.com/thefabric-io/specifications"
)

type Visitor struct {
	conditions   []string
	args         []interface{}
	fieldMap     map[string]string
	orderClauses []string
	limit        int
	offset       int
}

func NewVisitor(fieldMap map[string]string) *Visitor {
	return &Visitor{
		conditions:   []string{},
		args:         []interface{}{},
		fieldMap:     fieldMap,
		orderClauses: []string{},
		limit:        0,
		offset:       0,
	}
}

func (v *Visitor) mapField(domainField string) string {
	if dbField, ok := v.fieldMap[domainField]; ok {
		return dbField
	}
	return domainField
}

func (v *Visitor) VisitEqual(field string, value interface{}) {
	dbField := v.mapField(field)
	v.conditions = append(v.conditions, fmt.Sprintf("%s = ?", dbField))
	v.args = append(v.args, value)
}

func (v *Visitor) VisitIn(field string, values []interface{}) {
	dbField := v.mapField(field)
	if len(values) == 0 {
		v.conditions = append(v.conditions, "1=0")
		return
	}

	qs := make([]string, len(values))
	for i := range values {
		qs[i] = "?"
		v.args = append(v.args, values[i])
	}
	v.conditions = append(v.conditions, fmt.Sprintf("%s IN (%s)", dbField, strings.Join(qs, ", ")))
}

func (v *Visitor) VisitAnd(specs []specifications.Specification) {
	subVisitor := NewVisitor(v.fieldMap)

	for _, s := range specs {
		s.Accept(subVisitor)
	}

	if len(subVisitor.conditions) > 0 {
		v.conditions = append(v.conditions, "("+strings.Join(subVisitor.conditions, " AND ")+")")
		v.args = append(v.args, subVisitor.args...)
	}

	v.orderClauses = append(v.orderClauses, subVisitor.orderClauses...)

	if subVisitor.limit > 0 {
		v.limit = subVisitor.limit
	}

	if subVisitor.offset > 0 {
		v.offset = subVisitor.offset
	}
}

func (v *Visitor) VisitOr(specs []specifications.Specification) {
	subVisitor := NewVisitor(v.fieldMap)
	orParts := []string{}

	for _, s := range specs {
		temp := NewVisitor(v.fieldMap)

		s.Accept(temp)

		if len(temp.conditions) > 0 {
			orParts = append(orParts, "("+strings.Join(temp.conditions, " AND ")+")")
			subVisitor.args = append(subVisitor.args, temp.args...)
		}

		subVisitor.orderClauses = append(subVisitor.orderClauses, temp.orderClauses...)

		if temp.limit > 0 {
			subVisitor.limit = temp.limit
		}

		if temp.offset > 0 {
			subVisitor.offset = temp.offset
		}
	}

	if len(orParts) > 0 {
		v.conditions = append(v.conditions, "("+strings.Join(orParts, " OR ")+")")
		v.args = append(v.args, subVisitor.args...)
	}

	v.orderClauses = append(v.orderClauses, subVisitor.orderClauses...)

	if subVisitor.limit > 0 {
		v.limit = subVisitor.limit
	}

	if subVisitor.offset > 0 {
		v.offset = subVisitor.offset
	}
}

func (v *Visitor) VisitLimit(limit int) {
	v.limit = limit
}

func (v *Visitor) VisitOrder(field, direction string) {
	dbField := v.mapField(field)
	v.orderClauses = append(v.orderClauses, dbField+" "+direction)
}

func (v *Visitor) VisitGreaterThan(field string, value interface{}) {
	dbField := v.mapField(field)
	v.conditions = append(v.conditions, fmt.Sprintf("%s > ?", dbField))
	v.args = append(v.args, value)
}

func (v *Visitor) VisitLowerThan(field string, value interface{}) {
	dbField := v.mapField(field)
	v.conditions = append(v.conditions, fmt.Sprintf("%s < ?", dbField))
	v.args = append(v.args, value)
}

func (v *Visitor) VisitLike(field string, value interface{}) {
	dbField := v.mapField(field)
	// Typically LIKE patterns are expected to include '%' in the value
	v.conditions = append(v.conditions, fmt.Sprintf("%s LIKE ?", dbField))
	v.args = append(v.args, value)
}

func (v *Visitor) VisitOffset(offset int) {
	v.offset = offset
}

func (v *Visitor) VisitNotEqual(field string, value interface{}) {
	dbField := v.mapField(field)
	v.conditions = append(v.conditions, fmt.Sprintf("%s <> ?", dbField))
	v.args = append(v.args, value)
}

func (v *Visitor) VisitGreaterThanOrEqual(field string, value interface{}) {
	dbField := v.mapField(field)
	v.conditions = append(v.conditions, fmt.Sprintf("%s >= ?", dbField))
	v.args = append(v.args, value)
}

func (v *Visitor) VisitLowerThanOrEqual(field string, value interface{}) {
	dbField := v.mapField(field)
	v.conditions = append(v.conditions, fmt.Sprintf("%s <= ?", dbField))
	v.args = append(v.args, value)
}

func (v *Visitor) BuildQuery(baseQuery string) (string, []interface{}) {
	query := baseQuery
	if len(v.conditions) > 0 {
		fullCondition := strings.Join(v.conditions, " AND ")
		var finalQuery strings.Builder
		argIndex := 1
		for _, ch := range fullCondition {
			if ch == '?' {
				finalQuery.WriteString(fmt.Sprintf("$%d", argIndex))
				argIndex++
			} else {
				finalQuery.WriteRune(ch)
			}
		}
		query += " WHERE " + finalQuery.String()
	}

	if len(v.orderClauses) > 0 {
		query += " ORDER BY " + strings.Join(v.orderClauses, ", ")
	}

	if v.limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", v.limit)
	}
	if v.offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", v.offset)
	}

	return query, v.args
}
