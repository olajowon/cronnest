package database

import (
	"github.com/jinzhu/gorm"
	"fmt"
	"cronnest/configure"
)

var DB *gorm.DB
var err error

func init() {
	DB, err = gorm.Open("postgres", configure.PgSQL)
	if err != nil {
		fmt.Printf("db connect error %v", err)
	} else {
		fmt.Printf("db connect success %v", err)
	}

	if DB.Error != nil {
		fmt.Printf("database error %v", DB.Error)
	}

}