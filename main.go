package main

import (
	"fmt"
	"time"

	"github.com/Ravillatypov/callchecker/caller"
	"github.com/Ravillatypov/callchecker/database"
	_ "github.com/go-sql-driver/mysql"
	"github.com/ivahaev/amigo"
	"github.com/jmoiron/sqlx"
)

func main() {
	conf := amigo.Settings{Host: "192.168.20.50", Username: "admin", Password: "3798be401f476b788a26d84073f3abe2"}
	ami := amigo.New(&conf)
	ami.Connect()
	ami.On("connect", func(message string) {
		fmt.Println("Connected", message)
	})
	ami.On("error", func(message string) {
		fmt.Println("Connection error:", message)
		return
	})
	time.Sleep(time.Second)
	if db, err := sqlx.Connect("mysql", "root@/suzdev"); err != nil {
		fmt.Println(err.Error())
		return
	}
	d := database.DB{d: db}
	call2db := make(chan map[string]string)
	db2call := make(chan map[string]string)
	a := caller.Call{ami: ami, queue: 0}
	go d.Run(call2db, db2call)
	go a.Run(db2call, call2db)
}
