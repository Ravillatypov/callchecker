package main

import (
	"fmt"
	"time"

	"flag"

	"github.com/Ravillatypov/callchecker/caller"
	"github.com/Ravillatypov/callchecker/database"
	_ "github.com/go-sql-driver/mysql"
	"github.com/ivahaev/amigo"
	"github.com/jmoiron/sqlx"
)

func main() {
	var host, user, pass, sqlconf string
	var maxchan uint
	flag.StringVar(&host, "h", "localhost", "AMI hostname or ip address")
	flag.StringVar(&user, "u", "admin", "AMI username")
	flag.StringVar(&pass, "p", "amp111", "AMI secret")
	flag.StringVar(&sqlconf, "m", "root@/checker", "mysql configuration string as [username[:password]@][protocol[(address)]]/dbname")
	flag.UintVar(&maxchan, "c", 1, "number of active calls")
	flag.Parse()
	conf := amigo.Settings{Host: host, Username: user, Password: pass}
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
	db, err := sqlx.Connect("mysql", sqlconf)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	d := database.DB{}
	d.Init(db)
	call2db := make(chan map[string]string, 10)
	db2call := make(chan map[string]string, 10)
	a := caller.Call{}
	a.Init(ami, uint8(maxchan))
	go d.Run(call2db, db2call)
	a.Run(db2call, call2db)
}
