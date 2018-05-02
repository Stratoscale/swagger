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
}

type parseFn func(string) (interface{}, bool)

type filterField struct {
	exp   string
	parse parseFn
	wrap  WrapFn
}

// NewBuilder initialize a Builder and parse the passing Model that will be used in
// the Parse calls.
func NewBuilder(c *Config) (*Builder, error) {
	if err := c.defaults(); err != nil {
		return nil, err
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
			if field.IsEmbedded() {
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
		Sort:  b.DefaultSort,
		Limit: b.DefaultLimit,
	}
	// parse and validate limit.
	if v := params.Get(b.LimitParam); v != "" {
		n, err := parseNumber(b.LimitParam, v, 1, b.LimitMaxValue)
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

// parseField handle sort and filter fields.
func (b *Builder) parseField(field *structs.Field) {
	// get all options from the struct field.
	options := strings.Split(field.Tag(b.TagName), ",")
	// struct field has a sort option.
	if contains(options, sortTag) {
		b.sortFields[gorm.ToDBName(field.Name())] = true
	}
	// struct field has a filter option.
	if !contains(options, filterTag) {
		return
	}
	colName := gorm.ToDBName(field.Name())
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
		b.addStringField(colName, withSep, wrapFn)
	case int, *int, time.Time, *time.Time:
		parseFn := parseInt
		if _, ok := v.(time.Time); ok {
			parseFn = parseDate
		}
		if _, ok := v.(*time.Time); ok {
			parseFn = parseDatePointer
		}
		b.addFilterField(withSep+opEqual, colName+" = ?", parseFn)
		b.addFilterField(withSep+opNotEqual, colName+" <> ?", parseFn)
		b.addFilterField(withSep+opLessThan, colName+" < ?", parseFn)
		b.addFilterField(withSep+opLessThanOrEqual, colName+" <= ?", parseFn)
		b.addFilterField(withSep+opGreaterThan, colName+" > ?", parseFn)
		b.addFilterField(withSep+opGreaterThanOrEqual, colName+" >= ?", parseFn)
	default:
		typ := reflect.TypeOf(v)
		// add more cases if needed.
		if typ.ConvertibleTo(reflect.TypeOf([]string{})) {
			b.addStringField(colName, withSep, wrapFn)
		}
	}
}

// addStringField adds all string filters to the given field.
func (b *Builder) addStringField(colName, withSep string, wrap WrapFn) {
	b.addFilterField(withSep+opEqual, colName+" = ?", parseString, wrap)
	b.addFilterField(withSep+opNotEqual, colName+" <> ?", parseString, wrap)
	b.addFilterField(withSep+opLike, colName+" LIKE ?", parseLikeString, wrap)
}

// addFilterField gets field name, expression (format) and parse function, and
// add it to the filterFields.
func (b *Builder) addFilterField(name, format string, parse parseFn, wrap ...WrapFn) {
	wrapFn := nopWrapper
	if len(wrap) != 0 {
		wrapFn = wrap[0]
	}
	b.filterFields[name] = filterField{exp: format, parse: parse, wrap: wrapFn}
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
		if l[i] == s {
			return true
		}
	}
	return false
}

// filter parsers.
func parseInt(s string) (interface{}, bool) {
	n, err := strconv.Atoi(s)
	return s, err == nil && n >= 0
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
