package database

import (
	"github.com/jinzhu/gorm"
	"cronnest/configure"
	"log"
)

var DB *gorm.DB
var err error

func init() {
	DB, err = gorm.Open("postgres", configure.PgSQL)
	if err != nil {
		log.Fatalf("数据库连接失败, %v", err)
	}

	if DB.Error != nil {
		log.Fatalf("数据库错误, %v", DB.Error)
	}
}