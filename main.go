package main

import (
	"net/http"
	"time"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"cronnest/router"
	db "cronnest/database"
	"fmt"
)

func f() {
	fmt.Println("hheee")
}

func main() {
	defer db.DB.Close()

	router := router.InitRouter()
	s := &http.Server{
		Addr:           ":8090",
		Handler:        router,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	s.ListenAndServe()
}
