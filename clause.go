package sqlbuilder

import (
	"fmt"
	"strings"
)

// Clause represents a SQL where Clause
type Clause interface {
	// interpret interprets Clause into string
	interpret(sb *SelectBuilder) string
	// Not negatives Clause
	Not() *notClause
	// And connects several Clause into an andClause
	And(clause ...Clause) *andClause
	// Or connects several Clause into an orClause
	Or(clause ...Clause) *orClause
}

// newAndClause creates an *andClause
func newAndClause(augend Clause, addend ...Clause) *andClause {
	return &andClause{
		Augend: augend,
		Addend: addend,
	}
}

// andClause represents a SQL AND Clause
type andClause struct {
	Augend Clause
	Addend []Clause
}

func (a *andClause) interpret(sb *SelectBuilder) string {
	andExpr := make([]string, 0, len(a.Addend)+1)
	andExpr = append(andExpr, a.Augend.interpret(sb))
	for _, c := range a.Addend {
		andExpr = append(andExpr, c.interpret(sb))
	}
	return fmt.Sprintf("(%v)", strings.Join(andExpr, " AND "))
}

func (a *andClause) Not() *notClause {
	return newNotClause(a)
}

func (a *andClause) And(clause ...Clause) *andClause {
	return newAndClause(a, clause...)
}

func (a *andClause) Or(clause ...Clause) *orClause {
	return newOrClause(a, clause...)
}

// newOrClause creates an *orClause
func newOrClause(augend Clause, addend ...Clause) *orClause {
	return &orClause{
		Augend: augend,
		Addend: addend,
	}
}

// orClause represents a SQL OR Clause
type orClause struct {
	Augend Clause
	Addend []Clause
}

func (o *orClause) interpret(sb *SelectBuilder) string {
	orExpr := make([]string, 0, len(o.Addend)+1)
	orExpr = append(orExpr, o.Augend.interpret(sb))
	for _, c := range o.Addend {
		orExpr = append(orExpr, c.interpret(sb))
	}
	return fmt.Sprintf("(%v)", strings.Join(orExpr, " OR "))
}

func (o *orClause) Not() *notClause {
	return newNotClause(o)
}

func (o *orClause) And(clause ...Clause) *andClause {
	return newAndClause(o, clause...)
}

func (o *orClause) Or(clause ...Clause) *orClause {
	return newOrClause(o, clause...)
}

// newNotClause creates a notClause
func newNotClause(clause Clause) *notClause {
	return &notClause{
		clause,
	}
}

// notClause represents a SQL NOT Clause
type notClause struct {
	negend Clause
}

func (n *notClause) interpret(sb *SelectBuilder) string {
	return fmt.Sprintf("(NOT %v)", n.negend.interpret(sb))
}

func (n *notClause) Not() *notClause {
	return newNotClause(n)
}

func (n *notClause) And(clause ...Clause) *andClause {
	return newAndClause(n, clause...)
}

func (n *notClause) Or(clause ...Clause) *orClause {
	return newOrClause(n, clause...)
}

// basicClause represents a specific basic SQL where Clause
type basicClause struct {
	*operation
	operand []interface{}
}

func (b *basicClause) interpret(sb *SelectBuilder) string {
	return b.operate(sb, b.field, b.operand)
}

func (b *basicClause) Not() *notClause {
	return newNotClause(b)
}

func (b *basicClause) And(clause ...Clause) *andClause {
	return newAndClause(b, clause...)
}

func (b *basicClause) Or(clause ...Clause) *orClause {
	return newOrClause(b, clause...)
}

// operate interprets basicClause into string
type operate func(sb *SelectBuilder, field string, operand []interface{}) string

// newOperation creates an *operation
func newOperation(field string, operate operate) *operation {
	return &operation{
		field,
		operate,
	}
}

// operation stores field and operate of clause
type operation struct {
	field   string
	operate operate
}

// NewClause creates *basicClause with operand value
func (o *operation) NewClause(value ...interface{}) *basicClause {
	return &basicClause{
		o,
		value,
	}
}

// newZeroOperation creates a *zeroOperandOperation
func newZeroOperation(field string, operate operate) *zeroOperandOperation {
	return &zeroOperandOperation{
		newOperation(field, operate),
	}
}

// zeroOperandOperation can create *basicClause with zero operand
type zeroOperandOperation struct {
	*operation
}

// NewClause creates *basicClause with zero operand
func (z *zeroOperandOperation) NewClause() *basicClause {
	return z.operation.NewClause()
}

// newOneOperandOperation creates a *oneOperandOperation
func newOneOperandOperation(field string, operate operate) *oneOperandOperation {
	return &oneOperandOperation{
		newOperation(field, operate),
	}
}

// oneOperandOperation can create *basicClause with one operand
type oneOperandOperation struct {
	*operation
}

// NewClause creates *basicClause with one operand v
func (o *oneOperandOperation) NewClause(v interface{}) *basicClause {
	return o.operation.NewClause(v)
}

// newTwoOperandOperation creates a *twoOperandOperation
func newTwoOperandOperation(field string, operate operate) *twoOperandOperation {
	return &twoOperandOperation{
		newOperation(field, operate),
	}
}

// twoOperandOperation can create *basicClause with two operand
type twoOperandOperation struct {
	*operation
}

// NewClause creates *basicClause with operand v1, v2
func (t *twoOperandOperation) NewClause(v1, v2 interface{}) *basicClause {
	return t.operation.NewClause(v1, v2)
}

var (
	isNull operate = func(sb *SelectBuilder, field string, operand []interface{}) string {
		return sb.IsNull(field)
	}

	notNull operate = func(sb *SelectBuilder, field string, operand []interface{}) string {
		return sb.IsNotNull(field)
	}

	e operate = func(sb *SelectBuilder, field string, operand []interface{}) string {
		return sb.E(field, operand[0])
	}

	ne operate = func(sb *SelectBuilder, field string, operand []interface{}) string {
		return sb.NE(field, operand[0])
	}

	g operate = func(sb *SelectBuilder, field string, operand []interface{}) string {
		return sb.G(field, operand[0])
	}

	ge operate = func(sb *SelectBuilder, field string, operand []interface{}) string {
		return sb.GE(field, operand[0])
	}

	l operate = func(sb *SelectBuilder, field string, operand []interface{}) string {
		return sb.L(field, operand[0])
	}

	le operate = func(sb *SelectBuilder, field string, operand []interface{}) string {
		return sb.LE(field, operand[0])
	}

	like operate = func(sb *SelectBuilder, field string, operand []interface{}) string {
		return sb.Like(field, operand[0])
	}

	notLike operate = func(sb *SelectBuilder, field string, operand []interface{}) string {
		return sb.NotLike(field, operand[0])
	}

	between operate = func(sb *SelectBuilder, field string, operand []interface{}) string {
		return sb.Between(field, operand[0], operand[1])
	}

	notBetween operate = func(sb *SelectBuilder, field string, operand []interface{}) string {
		return sb.NotBetween(field, operand[0], operand[1])
	}

	in operate = func(sb *SelectBuilder, field string, operand []interface{}) string {
		return sb.In(field, operand...)
	}

	notIn operate = func(sb *SelectBuilder, field string, operand []interface{}) string {
		return sb.NotIn(field, operand...)
	}
)

// NewIsNullOperation creates a operation which can create Clause that represents "field IS NULL"
func NewIsNullOperation(field string) *zeroOperandOperation {
	return newZeroOperation(field, isNull)
}

// NewNotNullOperation creates operation which can create Clause that represents "field IS NOT NULL"
func NewNotNullOperation(field string) *zeroOperandOperation {
	return newZeroOperation(field, notNull)
}

// NewEqualOperation creates operation which can create Clause that represents "field = value"
func NewEqualOperation(field string) *oneOperandOperation {
	return newOneOperandOperation(field, e)
}

// NewNotEqualOperation creates operation which can create Clause that represents "field != value"
func NewNotEqualOperation(field string) *oneOperandOperation {
	return newOneOperandOperation(field, ne)
}

// NewGreaterThanOperation creates operation which can create Clause that represents "field > value"
func NewGreaterThanOperation(field string) *oneOperandOperation {
	return newOneOperandOperation(field, g)
}

// NewGreaterEqualThanOperation creates operation which can create Clause that represents "field >= value"
func NewGreaterEqualThanOperation(field string) *oneOperandOperation {
	return newOneOperandOperation(field, ge)
}

// NewLessThanOperation creates operation which can create Clause that represents "field < value"
func NewLessThanOperation(field string) *oneOperandOperation {
	return newOneOperandOperation(field, l)
}

// NewLessEqualThanOperation creates operation which can create Clause that represents "field <= value"
func NewLessEqualThanOperation(field string) *oneOperandOperation {
	return newOneOperandOperation(field, le)
}

// NewLikeOperation creates operation which can create Clause that represents "field LIKE value"
func NewLikeOperation(field string) *oneOperandOperation {
	return newOneOperandOperation(field, like)
}

// NewNotLikeOperation creates operation which can create Clause that represents "field NOT LIKE value"
func NewNotLikeOperation(field string) *oneOperandOperation {
	return newOneOperandOperation(field, notLike)
}

// NewBetweenOperation creates operation which can create Clause that represents "field BETWEEN lower AND upper"
func NewBetweenOperation(field string) *twoOperandOperation {
	return newTwoOperandOperation(field, between)
}

// NewNotBetweenOperation creates operation which can create Clause that represents "field NOT BETWEEN lower AND upper"
func NewNotBetweenOperation(field string) *twoOperandOperation {
	return newTwoOperandOperation(field, notBetween)
}

// NewInOperation creates operation which can create Clause that represents "field IN (value...)"
func NewInOperation(field string) *operation {
	return newOperation(field, in)
}

// NewNotInOperation creates operation which can create Clause that represents "field NOT IN (value...)"
func NewNotInOperation(field string) *operation {
	return newOperation(field, notIn)
}

// Interpret interprets Clause into string
func Interpret(clause Clause, sb *SelectBuilder) string {
	return clause.interpret(sb)
}
