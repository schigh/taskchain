package main

import (
	"crypto/rand"
	"encoding/hex"
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

func makeTask() func(*taskgraph.TaskGroup) error {
	return func(group *taskgraph.TaskGroup) error {

		logger, ok := group.Get("logger").(*log.Logger)
		if !ok {
			println("wat")
			return fmt.Errorf("unable to get logger")
		}

		sequence, ok := group.Get("sequence").(int)
		if !ok {
			return fmt.Errorf("unable to get sequence")
		}

		sequence++
		group.Set("sequence", sequence)

		b := make([]byte, 8)
		rand.Read(b)
		lbl := hex.EncodeToString(b)
		logger.Printf("SEQ: %d %s", sequence, lbl)
		if sequence%3 == 0 {
			return nil
		} else {
			return fmt.Errorf("ERROR: %s", lbl)
		}
	}
}

func main() {
	tg := &taskgraph.TaskGroup{
		Policies: taskgraph.HaltAfterTimeout,
	}
	tg2 := &taskgraph.TaskGroup{
		Policies: taskgraph.HaltOnAnyError | taskgraph.HaltAfterTimeout,
	}
	tg.Next = tg2
	tg2.Add(func(t *taskgraph.TaskGroup) error {
		lg.Printf("TASKGROUP 2")
		return nil
	})
	tg.Set("logger", lg)
	tg.Set("sequence", -1)
	for i := 0; i < 10; i++ {
		tg.Add(makeTask())
	}

	now := time.Now()
	err := tg.Exec()
	diff := time.Now().Sub(now)

	lg.Println(err)

	fmt.Printf("%dms\n", diff/1000000)
}
