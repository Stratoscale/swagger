package query

import "errors"

const (
	// fields in the struct tag.
	sortTag   = "sort"
	filterTag = "filter"
	paramTag  = "param"
	// search param in query string.
	searchParam = "search"
	// operators in query string.
	opEqual              = "eq"
	opNotEqual           = "neq"
	opLike               = "like"
	opLessThan           = "lt"
	opGreaterThan        = "gt"
	opLessThanOrEqual    = "lte"
	opGreaterThanOrEqual = "gte"
)

// An expression can be optionally prefixed with + or - to control the sorting direction,
// ascending or descending. For example, '+field' or '-field'.
// If the predicate is missing or empty then it defaults to '+'
var sortDirections = map[byte]string{'+': "asc", '-': "desc"}

// Config for the Builder constructor.
type Config struct {
	// Model is an instance of the struct definition. the Builder will parse
	// the url.Values according to this.
	Model interface{}
	// TagName is the name of the tag in the struct. defaults to "query".
	TagName string
	// Separator between field and command. defaults to "_".
	Separator string
	// IgnoreSort indicates if the builder should skip the sort process.
	IgnoreSort bool
	// SortParam is the name of the sort parameter.
	// defaults to "sort"
	SortParam string
	// DefaultSort is the default sort string for the query builder.
	// if the builder gets and empty sort parameter it'll add this default.
	DefaultSort string
	// LimitParam is the name of the limit parameter in the query string.
	// defaults to "limit".
	LimitParam string
	// DefaultLimit is the default value for limit option. default to 25.
	DefaultLimit int
	// LimitMaxValue is the maximum value that accept valid parameter.
	LimitMaxValue int
	// OffsetParam is the name of the offset parameter in the query string.
	// defaults to "offset"
	OffsetParam string
	// SearchOperator used to combine search condition together. defaults to "AND".
	SearchOperator string
}

func (c *Config) defaults() error {
	if c.Model == nil {
		return errors.New("query: 'Model' is a required field")
	}
	defaultString(&c.TagName, "query")
	defaultString(&c.Separator, "_")
	defaultString(&c.SortParam, "sort")
	defaultString(&c.LimitParam, "limit")
	defaultString(&c.OffsetParam, "offset")
	defaultString(&c.SearchOperator, "AND")
	defaultInt(&c.DefaultLimit, 25)
	defaultInt(&c.LimitMaxValue, 100)
	return nil
}

func defaultString(s *string, v string) {
	if *s == "" {
		*s = v
	}
}

func defaultInt(i *int, v int) {
	if *i == 0 {
		*i = v
	}
}
