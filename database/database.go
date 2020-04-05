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
		fmt.Printf("数据库连接失败, %v", err)
	}

	if DB.Error != nil {
		fmt.Printf("数据库错误, %v", DB.Error)
	}
}