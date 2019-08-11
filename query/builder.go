package query

import (
	"bytes"
	"container/list"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/structs"
	"github.com/jinzhu/gorm"
)

// Wrapper is the interface that wraps the wrap method.
type Wrapper interface {
	Wrap(string) string
}

// WrapFn is a function type that implements the Wrapper interface.
type WrapFn func(string) string

// Wrap is the function that implements the Wrapper interface.
// it called itself with the argument. nopWrapper is an example for it.
func (f WrapFn) Wrap(s string) string { return f(s) }

// nopWrapper is a wrapper that returns the string without changes.
// the reason for creating / using this instance, is to avoid nil checking.
// each field has one of the two: nopWrapper, or its own Wrapper implementation.
var nopWrapper = WrapFn(func(s string) string { return s })

// Searcher is the interface that wraps the Search method.
// Models that want to support search, need to implement this interface.
// if a "search" term is provided to the Parse method, the Builder will
// call the Search method with this value. the return value should be a search
// query for this term + its arguments. (we let gorm handle the escaping).
// If multiple search terms are provided, the Builder will use the SearchOperator to
// combine these search queries together.
type Searcher interface {
	Search(term string) (exp string, vals []interface{})
}

// ParseError is a typed error created dynamically based on the parsing failure.
type ParseError struct {
	msg string
}

// Error implements the error interface.
func (e ParseError) Error() string { return e.msg }

// Builder is a query builder.
// You should initialize it only once, and then use it in your http.Handler.
type Builder struct {
	*Config
	searcher     Searcher
	sortFields   map[string]bool
	filterFields map[string]filterField
	selectFields []string
}

type parseFn func(string) (interface{}, bool)

type filterField struct {
	exp          string
	parse        parseFn
	wrap         WrapFn
	splitOnComma bool
}

// NewBuilder initialize a Builder and parse the passing Model that will be used in
// the Parse calls.
func NewBuilder(c *Config) (*Builder, error) {
	if err := c.defaults(); err != nil {
		return nil, err
	}

	if c.OnlySelectNonDetailedFields {
		c.ExplicitSelect = true
	}

	b := &Builder{
		Config:       c,
		sortFields:   make(map[string]bool),
		filterFields: make(map[string]filterField),
	}
	if searcher, ok := c.Model.(Searcher); ok {
		b.searcher = searcher
	}
	b.init()
	return b, nil
}

// MustNewBuilder creates a new builder and panic on failure
func MustNewBuilder(c *Config) *Builder {
	b, err := NewBuilder(c)
	if err != nil {
		panic(err)
	}
	return b
}

// Move typ to config and comment that init should be called only once.
func (b *Builder) init() {
	// build the sort-fields and filter-fields data structures.
	l := list.New()
	l.PushFront(structs.Fields(b.Model))
	for l.Len() > 0 {
		fields := l.Remove(l.Front())
		for _, field := range fields.([]*structs.Field) {
			if field.IsEmbedded() && field.Kind() == reflect.Struct {
				l.PushFront(field.Fields())
				continue
			}
			b.parseField(field)
		}
	}
}

// Parse validates and parses the input params and return back a *DBQuery.
// It's safe to call it from multiple goroutines concurrently.
func (b *Builder) Parse(params url.Values) (*DBQuery, error) {
	q := &DBQuery{
		Sort:   b.DefaultSort,
		Limit:  b.DefaultLimit,
		Select: strings.Join(b.selectFields[:], ","),
	}
	// parse and validate limit.
	if v := params.Get(b.LimitParam); v != "" {
		n, err := parseNumber(b.LimitParam, v, 0, b.LimitMaxValue)
		if err != nil {
			return nil, err
		}
		q.Limit = n
	}
	// parse and validate offset.
	if v := params.Get(b.OffsetParam); v != "" {
		n, err := parseNumber(b.OffsetParam, v, 0, -1)
		if err != nil {
			return nil, err
		}
		q.Offset = n
	}
	// parse and validate sort parameters.
	if sortFields, ok := params[b.SortParam]; !b.IgnoreSort && ok {
		sortExp, err := b.parseSort(sortFields)
		if err != nil {
			return nil, err
		}
		q.Sort = sortExp
	}
	// parse and validate conditions and filter parameters.
	exp, val, err := b.parseFilter(params)
	if err != nil {
		return nil, err
	}
	q.CondExp, q.CondVal = exp, val
	// model implements the searcher interface.
	if terms, ok := params[searchParam]; ok && b.searcher != nil {
		exp, vals := b.parseSearch(terms)
		q.And(exp, vals...)
	}
	return q, nil
}

// ParseRequest is a helper function for parsing query from a request object
func (b *Builder) ParseRequest(r *http.Request) (*DBQuery, error) {
	return b.Parse(r.URL.Query())
}

// parseSearch generates search query for the given terms.
func (b *Builder) parseSearch(terms []string) (string, []interface{}) {
	var (
		vals []interface{}
		exp  = new(bytes.Buffer)
	)
	if len(terms) > 1 {
		exp.WriteString("(")
	}
	for i, term := range terms {
		tExp, tVals := b.searcher.Search(term)
		vals = append(vals, tVals...)
		exp.WriteString(tExp)
		if i != len(terms)-1 {
			exp.WriteString(fmt.Sprintf(" %s ", b.SearchOperator))
		}
	}
	if len(terms) > 1 {
		exp.WriteString(")")
	}
	return exp.String(), vals
}

// parseFilter builds condition expression and condition values from
// the given params based on the struct configuration.
func (b *Builder) parseFilter(params url.Values) (string, []interface{}, error) {
	var (
		filterExp []string
		filterVal []interface{}
	)
	for name, filter := range b.filterFields {
		args, ok := params[name]
		// ignore irrelevant fields
		if !ok {
			continue
		}
		if filter.splitOnComma && len(args) == 1 && strings.Contains(args[0], ",") {
			args = strings.Split(args[0], ",")
		}
		// there are two expression formats:
		// 1. "KEY = VAL"                     - when only one argument is given.
		// 2. "(KEY = VAL OR KEY = VAL2 ...)" - when multiple values are given.
		// we use "=" in this example, but it could be any other operator.
		expArgs := make([]string, 0, len(args))
		for _, arg := range args {
			v, ok := filter.parse(arg)
			if !ok {
				return "", nil, &ParseError{fmt.Sprintf("invalid parameter for key '%s'", name)}
			}
			filterVal = append(filterVal, v)
			// collect expressions.
			expArgs = append(expArgs, filter.exp)
		}
		// if there's more than one argument, concatenate with "OR".
		exp := strings.Join(expArgs, " OR ")
		if len(expArgs) > 1 {
			exp = "(" + exp + ")"
		}
		filterExp = append(filterExp, filter.wrap(exp))
	}
	return strings.Join(filterExp, " AND "), filterVal, nil
}

// parseSort builds a sort input for the DBQuery.
// sort param could be string with prefixed by '-', or '+' and
// an ordering indicator.
func (b *Builder) parseSort(fields []string) (string, error) {
	sortParams := make([]string, len(fields))
	for i, field := range fields {
		if field == "" {
			return "", &ParseError{"missing sort parameter"}
		}
		var orderBy string
		// if the sort field prefixed by order indicator
		if order, ok := sortDirections[field[0]]; ok {
			orderBy = order
			field = field[1:]
		}
		if !b.sortFields[field] {
			return "", &ParseError{fmt.Sprintf("invalid sort parameter '%s'", field)}
		}
		sortParams[i] = field
		if orderBy != "" {
			field += " " + orderBy
		}
		sortParams[i] = field
	}
	return strings.Join(sortParams, ", "), nil
}

// parse number. return an error if the string is invalid
// number and above/below the boundaries.
func parseNumber(k, v string, min, max int) (int, error) {
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, &ParseError{fmt.Sprintf("invalid value('%s') for key '%s'", v, k)}
	}
	if n < min {
		return 0, &ParseError{fmt.Sprintf("value for key '%s' must be greater than or equal to %d", k, min)}
	}
	if max != -1 && n > max {
		return 0, &ParseError{fmt.Sprintf("value for key '%s' must be less than or equal to %d", k, max)}
	}
	return n, nil
}

func (b *Builder) addFilterFieldsForNumericFields(withSep, colName string, parse parseFn, splitOnComma bool) {
	b.addFilterField(colName, colName+" = ?", parse, splitOnComma)
	b.addFilterField(withSep+opEqual, colName+" = ?", parse, splitOnComma)
	b.addFilterField(withSep+opNotEqual, colName+" <> ?", parse, splitOnComma)
	b.addFilterField(withSep+opLessThan, colName+" < ?", parse, splitOnComma)
	b.addFilterField(withSep+opLessThanOrEqual, colName+" <= ?", parse, splitOnComma)
	b.addFilterField(withSep+opGreaterThan, colName+" > ?", parse, splitOnComma)
	b.addFilterField(withSep+opGreaterThanOrEqual, colName+" >= ?", parse, splitOnComma)
}

func (b *Builder) addFilterFieldsForBoolFields(withSep, colName string, parse parseFn, splitOnComma bool) {
	b.addFilterField(colName, colName+" = ?", parse, splitOnComma)
	b.addFilterField(withSep+opEqual, colName+" = ?", parse, splitOnComma)
	b.addFilterField(withSep+opNotEqual, colName+" <> ?", parse, splitOnComma)
}

var (
	ignoreOptions []string = []string{
		"-",
		"foreignkey",
		"association_foreignkey",
		"many2many",
	}
)

func (b *Builder) appendToSelect(colName string, gormOptions []string, options []string) {
	for _, s := range ignoreOptions {
		if contains(gormOptions, s) {
			return
		}
	}
	if b.OnlySelectNonDetailedFields && contains(options, detailedTag) {
		return
	}

	b.selectFields = append(b.selectFields, colName)

}

// parseField handle sort and filter fields.
func (b *Builder) parseField(field *structs.Field) {
	colName := gorm.ToDBName(field.Name())

	// get all options from the struct field.
	options := strings.Split(field.Tag(b.TagName), ",")
	gormOptions := strings.Split(field.Tag("gorm"), ";")

	if b.ExplicitSelect {
		b.appendToSelect(colName, gormOptions, options)
	}

	// struct field has a sort option.
	if contains(options, sortTag) {
		b.sortFields[colName] = true
	}
	// struct field has a filter option.
	if !contains(options, filterTag) {
		return
	}

	splitOnComma := contains(options, splitTag)
	// if it has custom query-param, use it instead.
	if field, ok := hasQueryParam(options); ok {
		colName = field
	}
	var (
		v       = field.Value()
		wrapFn  = nopWrapper
		withSep = colName + b.Separator
	)
	// custom type may implements the Wrapper interface.
	if wrapper, ok := v.(Wrapper); ok {
		wrapFn = wrapper.Wrap
	}
	switch v.(type) {
	case string, *string:
		b.addStringField(colName, withSep, splitOnComma, wrapFn)
	case int, *int:
		parseFn := parseInt
		b.addFilterFieldsForNumericFields(withSep, colName, parseFn, splitOnComma)
	case int64, *int64:
		parseFn := parseInt64
		b.addFilterFieldsForNumericFields(withSep, colName, parseFn, splitOnComma)
	case time.Time:
		parseFn := parseDate
		b.addFilterFieldsForNumericFields(withSep, colName, parseFn, splitOnComma)
	case *time.Time:
		parseFn := parseDatePointer
		b.addFilterFieldsForNumericFields(withSep, colName, parseFn, splitOnComma)
	case bool, *bool:
		parseFn := parseBool
		b.addFilterFieldsForBoolFields(withSep, colName, parseFn, splitOnComma)
	default:
		typ := reflect.TypeOf(v)
		_, isStringer := v.(fmt.Stringer)

		// add more cases if needed.
		dummyString := ""
		switch {
		case typ.ConvertibleTo(reflect.TypeOf(dummyString)), typ.ConvertibleTo(reflect.TypeOf(&dummyString)):
			b.addStringField(colName, withSep, splitOnComma, wrapFn)
		case typ.ConvertibleTo(reflect.TypeOf([]string{})):
			b.addStringField(colName, withSep, splitOnComma, wrapFn)
		case isStringer:
			b.addStringField(colName, withSep, splitOnComma, wrapFn)
		default:
			panic(fmt.Sprintf("Could not use field %s (%T) with query filter", field.Name(), v))
		}

	}
}

// addStringField adds all string filters to the given field.
func (b *Builder) addStringField(colName, withSep string, splitOnComma bool, wrap WrapFn) {
	b.addFilterField(colName, colName+" = ?", parseString, splitOnComma, wrap)
	b.addFilterField(withSep+opEqual, colName+" = ?", parseString, splitOnComma, wrap)
	b.addFilterField(withSep+opNotEqual, colName+" <> ?", parseString, splitOnComma, wrap)
	b.addFilterField(withSep+opLike, colName+" LIKE ?", parseLikeString, splitOnComma, wrap)
}

// addFilterField gets field name, expression (format) and parse function, and
// add it to the filterFields.
func (b *Builder) addFilterField(name, format string, parse parseFn, splitOnComma bool, wrap ...WrapFn) {
	wrapFn := nopWrapper
	if len(wrap) != 0 {
		wrapFn = wrap[0]
	}
	b.filterFields[name] = filterField{exp: format, parse: parse, wrap: wrapFn, splitOnComma: splitOnComma}
}

// hasQueryParam return the custom param if there is one.
func hasQueryParam(l []string) (string, bool) {
	for _, s := range l {
		if strings.HasPrefix(s, paramTag) {
			return strings.TrimPrefix(s, paramTag+"="), true
		}
	}
	return "", false
}

// contains test if string is in the given list.
func contains(l []string, s string) bool {
	for i := range l {
		if l[i] == s || strings.HasPrefix(l[i], s) {
			return true
		}
	}
	return false
}

// filter parsers.
func parseInt(s string) (interface{}, bool) {
	if s == "" {
		return nil, false
	}
	n, err := strconv.Atoi(s)
	return n, err == nil
}

func parseInt64(s string) (interface{}, bool) {
	if s == "" {
		return nil, false
	}
	n, err := strconv.ParseInt(s, 10, 64)
	return n, err == nil
}

func parseString(s string) (interface{}, bool) {
	return s, s != ""
}

func parseLikeString(s string) (interface{}, bool) {
	return "%" + s + "%", s != ""
}

func parseDate(s string) (interface{}, bool) {
	t, err := time.Parse(time.RFC3339, s)
	return t, err == nil
}

func parseDatePointer(s string) (interface{}, bool) {
	t, ok := parseDate(s)
	if !ok {
		return nil, false
	}
	return &t, true
}

func parseBool(s string) (interface{}, bool) {
	if s == "" {
		return nil, false
	}
	_, err := strconv.ParseBool(s)
	if err != nil {
		return nil, false
	}
	return s, true
}
