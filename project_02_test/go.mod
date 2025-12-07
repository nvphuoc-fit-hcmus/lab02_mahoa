module project_02_test

go 1.25.4

require (
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/stretchr/testify v1.11.1
	golang.org/x/crypto v0.33.0
	gorm.io/gorm v1.25.5
	lab02_mahoa v0.0.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.17 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/driver/sqlite v1.5.4 // indirect
)

replace lab02_mahoa => ../project_02_source
