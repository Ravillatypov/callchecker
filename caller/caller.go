package caller

import (
	"fmt"
	"time"

	"strconv"

	"github.com/ivahaev/amigo"
)

// Call структура хранит указатель на ami и число запущенных звонков
type Call struct {
	queue uint8
	ami   *amigo.Amigo
}

// Init для инициализации
func (c *Call) Init(a *amigo.Amigo) {
	c.queue = 0
	c.ami = a
}

// Run метод для запуска
func (c *Call) Run(in, out chan map[string]string) {
	go c.getresult(out)
	if c.queue == 0 {
		out <- map[string]string{
			"type": "select",
		}
	}
	for a := range in {
		if count, ok := strconv.Atoi(a["count"]); ok == nil && count < 1 {
			fmt.Println("всех обзвонили, следующий обзвон через 40 мин")
			time.Sleep(time.Second * 2400)
		}
		if phone, ok := a["phone"]; ok {
			c.runcall(phone)
		}
	}
}
func (c *Call) runcall(number string) {
	if c.queue < 80 && c.ami.Connected() {
		c.ami.Action(map[string]string{
			"Action":   "Originate",
			"Channel":  "local/" + number + "@call-checker",
			"Context":  "call-checker",
			"Exten":    "88008008080",
			"Priority": "1",
			"Callerid": "1003",
			"Timeout":  "30000",
			"Async":    "true",
		})
		c.queue++
	} else {
		time.Sleep(time.Second * 10)
		c.runcall(number)
	}
}

func (c *Call) getresult(out chan map[string]string) {
	ch := make(chan map[string]string, 100)
	c.ami.SetEventChannel(ch)

	for e := range ch {
		if e["Event"] == "DialEnd" && e["Context"] == "call-checker" {
			if len(e["DestExten"]) < 9 || e["DestExten"] == "88008008080" {
				continue
			}
			out <- map[string]string{
				"type":   "update",
				"phone":  e["DestExten"],
				"status": e["DialStatus"],
			}
			c.queue--
		}
	}
}
