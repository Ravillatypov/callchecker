package caller

import (
	"fmt"
	"time"

	"os"
	"sync"

	"github.com/ivahaev/amigo"
)

// Call структура хранит указатель на ami и число запущенных звонков
type Call struct {
	queue uint8
	m     sync.Map
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
		if _, ok := a["end"]; ok {
			out <- map[string]string{
				"type": "select",
			}
		}
		if _, ok := a["stop"]; ok {
			fmt.Println("exit...")
			os.Exit(0)
		}
		if count, ok := a["count"]; ok && count == "0" {
			fmt.Println("всех обзвонили, ждем 30 мин")
			time.Sleep(time.Minute * 30)
		}
		if phone, ok := a["phone"]; ok {
			go c.runcall(phone, 1)
		}
	}
}
func (c *Call) runcall(number string, sleep int) {
	if sleep < 2 {
		sleep = 2
	}
	if c.queue < 10 && c.ami.Connected() {
		time.Sleep(time.Second * time.Duration(sleep))
		c.ami.Action(map[string]string{
			"Action":   "Originate",
			"Channel":  "local/" + number + "@call-checker",
			"Context":  "call-checker",
			"Exten":    "88008008080",
			"Priority": "1",
			"Callerid": "1011",
			"Timeout":  "30000",
			"Async":    "true",
		})
		c.m.Store(number, number)
		c.queue++
		fmt.Printf("queue: %d\n", c.queue)
		time.Sleep(time.Second * 31)
		if _, ok := c.m.Load(number); ok {
			c.queue--
			c.m.Delete(number)
			fmt.Printf("queue: %d skip number %s\n", c.queue, number)
		}
	} else {
		time.Sleep(time.Duration(sleep*2) * time.Second)
		c.runcall(number, 5)
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
			ph, _ := e["DestExten"]
			if _, ok := c.m.Load(ph); ok {
				out <- map[string]string{
					"type":   "update",
					"phone":  ph,
					"status": e["DialStatus"],
				}
				c.queue--
				fmt.Printf("queue: %d\n", c.queue)
				c.m.Delete(ph)
			}
		}
		if e["Event"] == "VarSet" && e["Context"] == "call-checker" && (e["Variable"] == "CALLSTATUS" || e["Variable"] == "DIALSTATUS") {
			ph, _ := e["Exten"]
			st, _ := e["Value"]
			if len(st) > 2 && ph != "88008008080" && ph != "1011" {

				if _, ok := c.m.Load(ph); ok {
					out <- map[string]string{
						"type":   "update",
						"phone":  ph,
						"status": st,
					}
					c.queue--
					fmt.Printf("queue2: %d\n", c.queue)
					c.m.Delete(ph)
				}
			}
		}
	}
}
