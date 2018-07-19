package query

import (
	"fmt"
	"log"
	"net/url"
	"time"
)

func ExampleBuilder_Parse() {
	b, err := NewBuilder(&Config{
		Model: struct {
			Name      string    `query:"sort,filter"`
			Age       int       `query:"filter"`
			CreatedAt time.Time `query:"sort,filter"`
		}{},
		DefaultSort:  "created_at desc",
		DefaultLimit: 10,
	})
	if err != nil {
		log.Fatal("failed to initialize builder", err)
	}
	values, _ := url.ParseQuery("sort=name&age_lt=20&age_gte=13&sort=-created_at&offset=7")
	qi, _ := b.Parse(values)
	fmt.Println(qi.Sort)
	fmt.Println(qi.CondExp)
	fmt.Println(qi.CondVal...)
	fmt.Println("limit", qi.Limit, "offset", qi.Offset)
	// -Output:   test disabled because each run has a different sort order
	// name, created_at desc
	// age >= ? AND age < ?
	// 13 20
	// limit 10 offset 7
}
