package query

import (
	"github.com/jinzhu/gorm"
)

// DBQuery are options for query a database
type DBQuery struct {
	// the number of rows returned by the SELECT statement.
	Limit int
	// start querying from offset x. used for pagination.
	Offset int
	// used as a parameter for the gorm.Order method. example: "age desc, name"
	Sort string
	// CondExp and CondVal come together and used as a parameters for the gorm.Where
	// method.
	//
	// examples:
	// 	1. Exp: "name = ?"
	//	   Val: "a8m"
	//
	//	2. Exp: "id IN (?)"
	//	   Val: []int{1,2,3}
	//
	//	3. Exp: "name = ? AND age >= ?"
	// 	   Val: "a8m", 22
	CondExp string
	CondVal []interface{}
	// Select specify fields that you want to retrieve from database when querying.
	// the default is to select all fields.
	//
	//	Select: "DISTINCT id"
	Select string
}

// Apply applies the query input on a database instance
func (q *DBQuery) Apply(db *gorm.DB) *gorm.DB {
	if q == nil {
		return db
	}
	if q.Offset != 0 {
		db = db.Offset(q.Offset)
	}
	if q.Limit != 0 {
		db = db.Limit(q.Limit)
	}
	if q.Select != "" {
		db = db.Select(q.Select)
	}
	if q.Sort != "" {
		db = db.Order(q.Sort)
	}
	if q.CondExp != "" {
		db = db.Where(q.CondExp, q.CondVal...)
	}
	return db
}

// And adds expression to the current where statement with AND condition
func (q *DBQuery) And(exp string, vals ...interface{}) {
	if q.CondExp != "" {
		q.CondExp += " AND " // combine the two expressions together.
	}
	q.CondExp += exp
	q.CondVal = append(q.CondVal, vals...)
}
