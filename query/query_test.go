package query

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type MyEnum string

const (
	enumVal1 MyEnum = "v1"
	enumVal2 MyEnum = "v2"
)

type Tags []string

func (Tags) Wrap(s string) string {
	return fmt.Sprintf("(name IN (SELECT DISTINCT tag_name IN tags WHERE %s))", s)
}

// model used in the unit tests.
type model struct {
	Name           string    `query:"sort,filter"`
	Status         string    `query:"filter"`
	Age            int64     `query:"filter"`
	Year           int       `query:"filter,detailed"`
	Dummy          int       `gorm:"-"`
	CreatedAt      time.Time `query:"sort,filter"`
	UpdatedAt      time.Time `query:"sort,filter"`
	Tags           Tags      `query:"filter,param=tag_name"`
	FlagPtr        *bool     `query:"filter,sort"`
	Flag           bool      `query:"filter,sort"`
	EnumValPtr     *MyEnum   `query:"filter,sort"`
	EnumVal        MyEnum    `query:"filter,sort"`
	FieldToIgnore1 bool      `gorm:"foreignkey:ExternalEndpointID"`
	FieldToIgnore2 bool      `gorm:"association_foreignkey:UUID"`
	FieldToIgnore3 bool      `gorm:"many2many:UUID"`
}

const (
	detailedFields    = "name,status,age,year,created_at,updated_at,tags,flag_ptr,flag,enum_val_ptr,enum_val"
	nonDetailedFields = "name,status,age,created_at,updated_at,tags,flag_ptr,flag,enum_val_ptr,enum_val"
)

// Search is an implementation of Seacrher used for testing.
func (model) Search(val string) (string, []interface{}) {
	return fmt.Sprintf("(name = ? OR status LIKE ?)"), []interface{}{val, "%" + val + "%"}
}

func TestQuery(t *testing.T) {
	var testCases = []struct {
		name string

		configInput *Config
		parseInput  url.Values

		expectedParseError error
		expectedQueryInput *DBQuery
	}{
		{
			name:       "simple example with default values",
			parseInput: url.Values{},
			configInput: &Config{
				DefaultSort: "name desc",
			},
			expectedQueryInput: &DBQuery{
				Limit:  25,
				Offset: 0,
				Sort:   "name desc",
			},
		},
		{
			name: "multiple sort options. with sort directions.",
			parseInput: url.Values{
				"sort": []string{"+updated_at", "name", "-created_at"},
			},
			configInput: new(Config),
			expectedQueryInput: &DBQuery{
				Limit:  25,
				Offset: 0,
				Sort:   "name desc",
			},
		},
		{
			name: "simple filters",
			parseInput: url.Values{
				"flag":         []string{"true"},
				"flag_ptr_eq":  []string{"false"},
				"enum_val":     []string{"v1"},
				"enum_val_ptr": []string{"v2"},
				"age_gt":       []string{"10"},
				"name_eq":      []string{"a8m", "pos", "yossi"},
			},
			configInput: &Config{},
			expectedQueryInput: &DBQuery{
				Limit:   25,
				Offset:  0,
				CondExp: "age > ? AND (name = ? OR name = ? OR name = ?) AND flag = ? AND flag_ptr = ? AND enum_val = ? AND enum_val_ptr = ?",
				CondVal: []interface{}{int64(10), "a8m", "pos", "yossi", "true", "false", "v1", "v2"},
			},
		},
		{
			name: "multiple filters. with custom configuration",
			parseInput: url.Values{
				"lp":      []string{"10"},
				"oft":     []string{"5"},
				"age_gt":  []string{"10"},
				"name_eq": []string{"a8m", "pos", "yossi"},
			},
			configInput: &Config{
				DefaultLimit: 100,
				OffsetParam:  "oft",
				LimitParam:   "lp",
			},
			expectedQueryInput: &DBQuery{
				Limit:   10,
				Offset:  5,
				CondExp: "age > ? AND (name = ? OR name = ? OR name = ?)",
				CondVal: []interface{}{int64(10), "a8m", "pos", "yossi"},
			},
		},
		{
			name: "multiple filters. with exact names",
			parseInput: url.Values{
				"lp":     []string{"10"},
				"oft":    []string{"5"},
				"age_gt": []string{"10"},
				"name":   []string{"a8m", "pos", "yossi"},
			},
			configInput: &Config{
				DefaultLimit: 100,
				OffsetParam:  "oft",
				LimitParam:   "lp",
			},
			expectedQueryInput: &DBQuery{
				Limit:   10,
				Offset:  5,
				CondExp: "age > ? AND (name = ? OR name = ? OR name = ?)",
				CondVal: []interface{}{int64(10), "a8m", "pos", "yossi"},
			},
		},
		{
			name: "custom tag name",
			configInput: &Config{
				TagName: "ormquery",
				Model: struct {
					Name string `ormquery:"sort,filter"`
				}{},
			},
			parseInput: url.Values{
				"name_like": []string{"a8m"},
			},
			expectedQueryInput: &DBQuery{
				Limit:   25,
				CondExp: "name LIKE ?",
				CondVal: []interface{}{"%a8m%"},
			},
		},
		{
			name: "expression delegation",
			configInput: &Config{
				Model: &model{},
			},
			parseInput: url.Values{
				"tag_name_like": []string{"a8m"},
			},
			expectedQueryInput: &DBQuery{
				Limit:   25,
				CondExp: "(name IN (SELECT DISTINCT tag_name IN tags WHERE tag_name LIKE ?))",
				CondVal: []interface{}{"%a8m%"},
			},
		},
		{
			name: "expression delegation",
			configInput: &Config{
				Model: &model{},
			},
			parseInput: url.Values{
				"age_gte":     []string{"27"},
				"tag_name_eq": []string{"a", "b"},
			},
			expectedQueryInput: &DBQuery{
				Limit:   25,
				CondExp: "age >= ? AND (name IN (SELECT DISTINCT tag_name IN tags WHERE (tag_name = ? OR tag_name = ?)))",
				CondVal: []interface{}{"a", "b"},
			},
		},
		{
			name: "one search term",
			configInput: &Config{
				Model: &model{},
			},
			parseInput: url.Values{
				"search": []string{"foo"},
			},
			expectedQueryInput: &DBQuery{
				Limit:   25,
				CondExp: "(name = ? OR status LIKE ?)",
				CondVal: []interface{}{"foo", "%foo%"},
			},
		},
		{
			name: "multiple search term",
			configInput: &Config{
				Model: &model{},
			},
			parseInput: url.Values{
				"search": []string{"foo", "bar"},
			},
			expectedQueryInput: &DBQuery{
				Limit:   25,
				CondExp: "((name = ? OR status LIKE ?) AND (name = ? OR status LIKE ?))",
				CondVal: []interface{}{"foo", "%foo%", "bar", "%bar%"},
			},
		},
		{
			name: "multiple search term with another filter",
			configInput: &Config{
				Model: &model{},
			},
			parseInput: url.Values{
				"search":  []string{"foo", "bar"},
				"age_eq":  []string{"20"},
				"year_eq": []string{"1998"},
			},
			expectedQueryInput: &DBQuery{
				Limit:   25,
				CondExp: "age = ? AND year = ? AND ((name = ? OR status LIKE ?) AND (name = ? OR status LIKE ?))",
				CondVal: []interface{}{"foo", "%foo%", "bar", "%bar%", int64(20), 1998},
			},
		},
		{
			name: "select with details and explicit",
			parseInput: url.Values{
				"sort": []string{"+updated_at", "name", "-created_at"},
			},
			configInput: &Config{
				Model:                       &model{},
				OnlySelectNonDetailedFields: false,
				ExplicitSelect: true,
			},
			expectedQueryInput: &DBQuery{
				Limit:  25,
				Offset: 0,
				Sort:   "name desc",
				Select: detailedFields,
			},
		},
		{
			name: "select with details",
			parseInput: url.Values{
				"sort": []string{"+updated_at", "name", "-created_at"},
			},
			configInput: &Config{
				Model:                       &model{},
				OnlySelectNonDetailedFields: false,
			},
			expectedQueryInput: &DBQuery{
				Limit:  25,
				Offset: 0,
				Sort:   "name desc",
				Select: "",
			},
		},
		{
			name: "select with details by default",
			parseInput: url.Values{
				"sort": []string{"+updated_at", "name", "-created_at"},
			},
			configInput: &Config{
				Model: &model{},
			},
			expectedQueryInput: &DBQuery{
				Limit:  25,
				Offset: 0,
				Sort:   "name desc",
				Select: "",
			},
		},
		{
			name: "select without details",
			parseInput: url.Values{
				"sort": []string{"+updated_at", "name", "-created_at"},
			},
			configInput: &Config{
				Model:                       &model{},
				OnlySelectNonDetailedFields: true,
			},
			expectedQueryInput: &DBQuery{
				Limit:  25,
				Offset: 0,
				Sort:   "name desc",
				Select: nonDetailedFields,
			},
		},
	}
	for _, test := range testCases {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			if test.configInput.Model == nil {
				test.configInput.Model = model{}
			}
			builder, _ := NewBuilder(test.configInput)
			qi, err := builder.Parse(test.parseInput)
			if err != nil {
				assert.IsType(t, test.expectedParseError, err, err.Error())
				return
			}
			assert.Equal(t, test.expectedQueryInput.Limit, qi.Limit, "limit field")
			assert.Equal(t, test.expectedQueryInput.Offset, qi.Offset, "offset field")
			if test.expectedQueryInput.Select != "" {
				assert.Equal(t, test.expectedQueryInput.Select, qi.Select, "select field")
			}
			// use this technique because the order of the map iterations is unpredictable.
			// test that the expression is "equal".
			actual, expected := strings.Split(qi.CondExp, " AND "), strings.Split(test.expectedQueryInput.CondExp, " AND ")
			sort.Strings(actual)
			sort.Strings(expected)
			assert.Equal(t, actual, expected, "condition expression")
			// do the same thing for the condition values
			for _, val := range test.expectedQueryInput.CondVal {
				assert.Contains(t, qi.CondVal, val, "condition value")
			}
		})
	}
}
