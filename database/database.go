package database

import (
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
	db.d.Select(&res, `SELECT phone FROM contact_phones
	WHERE status != 'ANSWER' AND lastcall < NOW()-10000 AND count < 5
	LIMIT 1000`)
	return res
}
func (db *DB) updateNumber(phone, status string) {
	db.d.Exec(`UPDATE contact_phones SET count=count+1,lastcall=NOW(),status=? 
	WHERE phone=?`, status, phone)
}

// Run публичный метод для запуска
func (db *DB) Run(in, out chan map[string]string) {
	for a := range in {
		switch a["type"] {
		case "select":
			go func() {
				phs := db.getnumbers()
				for _, ph := range phs {
					out <- map[string]string{"phone": ph}
				}
				ln := string(len(phs))
				out <- map[string]string{"count": ln}
			}()
		case "update":
			go db.updateNumber(a["phone"], a["status"])
		}
	}
}
