module github.com/Stratoscale/swagger/testdata

go 1.12

replace github.com/Stratoscale/swagger/models => /home/ronnie/wgo1/src/github.com/Stratoscale/swagger/testdata/models

replace github.com/Stratoscale/swagger/testdat/auth => /home/ronnie/wgo1/src/github.com/Stratoscale/swagger/testdata/auth

replace github.com/Stratoscale/swagger => /home/ronnie/wgo1/src/github.com/Stratoscale/swagger

require (
	github.com/Stratoscale/swagger v1.0.27
	github.com/go-openapi/errors v0.19.2
	github.com/go-openapi/loads v0.19.3
	github.com/go-openapi/runtime v0.19.6
	github.com/go-openapi/spec v0.19.3
	github.com/go-openapi/strfmt v0.19.3
	github.com/go-openapi/swag v0.19.5
	github.com/go-openapi/validate v0.19.3
	github.com/jinzhu/gorm v1.9.11 // indirect
	github.com/stretchr/testify v1.4.0
	golang.org/x/net v0.0.0-20191002035440-2ec189313ef0
)
