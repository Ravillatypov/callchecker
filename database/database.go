package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// DB структура для хранения указателя к БД
type DB struct {
	d *sqlx.DB
}

// Init для инициализации
func (db *DB) Init(m *sqlx.DB) {
	db.d = m
}
func (db *DB) getnumbers() []string {
	var res []string
	err := db.d.Select(&res, `SELECT phone FROM suz_contact_phones
	WHERE status != 'ANSWER' AND lastcall < NOW()-10000 AND count < 5
	ORDER BY count LIMIT 10`)
	if err != nil {
		fmt.Println(err.Error())
	}
	return res
}
func (db *DB) getcount() uint16 {
	var res uint16
	err := db.d.QueryRowx(`SELECT COUNT(1) AS result FROM suz_contact_phones
	WHERE status != 'ANSWER' AND count < 5 LIMIT 1`).Scan(&res)
	if err != nil {
		fmt.Println(err.Error())
	}
	return res
}
func (db *DB) updateNumber(phone, status string) {
	_, err := db.d.Exec(`UPDATE suz_contact_phones SET count=count+1,lastcall=NOW(),status=? 
	WHERE phone=?`, status, phone)
	if err != nil {
		fmt.Println(err.Error())
	}
}

// Run публичный метод для запуска
func (db *DB) Run(in, out chan map[string]string) {
	for a := range in {
		switch a["type"] {
		case "select":
			go func() {
				phs := db.getnumbers()
				phonescount := db.getcount()
				for _, ph := range phs {
					out <- map[string]string{"phone": ph}
				}
				if len(phs) > 0 {
					out <- map[string]string{"end": "end"}
				} else {
					out <- map[string]string{"count": "0"}
				}
				if phonescount == 0 {
					out <- map[string]string{"stop": "stop"}
				}

			}()
		case "update":
			go db.updateNumber(a["phone"], a["status"])
			fmt.Printf("updated: %s\t\t%s\n", a["phone"], a["status"])
		}
	}
}
