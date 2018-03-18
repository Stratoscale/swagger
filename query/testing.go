package query

import (
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// AssertQueryEqual tests if two query input are equal
func AssertQueryEqual(t *testing.T, got *DBQuery, want *DBQuery) {
	assert.Equal(t, got.Limit, want.Limit, "limit field")
	assert.Equal(t, got.Offset, want.Offset, "offset field")
	// test that the expression is "equal"
	wantCond := strings.Split(got.CondExp, " AND ")
	sort.Strings(wantCond)
	gotCond := strings.Split(want.CondExp, " AND ")
	sort.Strings(gotCond)
	assert.Equal(t, wantCond, gotCond, "condition expression")
	// do the same thing for the condition values
	for _, val := range want.CondVal {
		assert.Contains(t, got.CondVal, val, "condition value")
	}
}
