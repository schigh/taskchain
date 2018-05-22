package main

import (
	"fmt"
	"github.com/schigh/taskgraph"
	"log"
	mrand "math/rand"
	"os"
	"sync"
	"time"
)

type FooType struct {
	ID   int
	Name string
}

type SafeFooType struct {
	sync.Mutex
	ID   int
	Name string
}

var (
	lg *log.Logger
)

func init() {
	lg = log.New(os.Stderr, "👉 ", 0)
	mrand.Seed(time.Now().UnixNano())
}

func main() {
	tg := &taskgraph.TaskGroup{}
	tg.Add(func(t *taskgraph.TaskGroup) error {
		lg := t.Get("logger", nil).(*log.Logger)
		lg.Println("TASKGROUP 1")

		tgg1 := &taskgraph.TaskGroup{}
		tgg1.Set("lawger", lg)
		tgg1.Add(func(tt *taskgraph.TaskGroup) error {
			lgg := tt.Get("lawger", nil).(*log.Logger)
			lgg.Println("\tTASKGROUP 1.1")
			return nil
		})

		tgg1.Add(func(tt *taskgraph.TaskGroup) error {
			lgg := tt.Get("lawger", nil).(*log.Logger)
			lgg.Println("\tTASKGROUP 1.2")

			tggg1 := &taskgraph.TaskGroup{}
			tggg1.Set("looger", lgg)
			tggg1.Add(func(ttt *taskgraph.TaskGroup) error {
				lggg := ttt.Get("looger", nil).(*log.Logger)
				lggg.Println("\t\tTASKGROUP 1.2.1")
				return nil
			})
			return tggg1.Exec()
		})

		tgg1.Add(func(tt *taskgraph.TaskGroup) error {
			lgg := tt.Get("lawger", nil).(*log.Logger)
			lgg.Println("\tTASKGROUP 1.3")
			return nil
		})

		tgg2 := &taskgraph.TaskGroup{}
		tgg1.Next = tgg2

		tgg2.Add(func(tt *taskgraph.TaskGroup) error {
			lgg := tt.Get("lawger", nil).(*log.Logger)
			lgg.Println("\tTASKGROUP 1.4.1")
			return nil
		})

		return tgg1.Exec()
	})

	tg2 := &taskgraph.TaskGroup{}
	tg.Next = tg2
	tg2.Add(func(t *taskgraph.TaskGroup) error {
		lg := t.Get("logger", nil).(*log.Logger)
		lg.Println("TASKGROUP 2")
		return nil
	})
	tg.Set("logger", lg)

	now := time.Now()
	err := tg.Exec()
	diff := time.Now().Sub(now)

	lg.Println(err)

	fmt.Printf("%dms\n", diff/1000000)
}
