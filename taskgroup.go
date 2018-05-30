package taskgraph

import "sync"

// TaskGroup is a grouping of Tasks that will be dispatched
// asynchronously
type TaskGroup struct {
	// Next is the task group that is to be executed after all
	// tasks within the current task group are completed
	Next *TaskGroup

	// ErrorHandler will handle all errors returned within this
	// task group
	ErrorHandler func(*TaskGroup, error)

	tasks []Task
	bag   *bag
}

// Add a Task to this group
func (t *TaskGroup) Add(task Task) {
	t.tasks = append(t.tasks, task)
}

// Exec will run the task group
func (t *TaskGroup) Exec() error {
	numTasks := len(t.tasks)
	errchan := make(chan error)
	donechan := make(chan struct{}, numTasks)
	defer func() {
		close(errchan)
		close(donechan)
	}()

	for _, td := range t.tasks {
		go func(task Task, grErrChan chan error, grDoneChan chan struct{}) {
			err := task(t)
			if err != nil {
				grErrChan <- err
			}
			grDoneChan <- struct{}{}
		}(td, errchan, donechan)
	}

	var errorOut error
	var firstErrorToken sync.Once
	pops := 0
	for {
		if pops >= numTasks {
			break
		}
		select {
		case <-donechan:
			pops++
		case err := <-errchan:
			firstErrorToken.Do(func() {
				errorOut = err
			})
			if t.ErrorHandler != nil {
				go t.ErrorHandler(t, err)
			}
		}
	}

	if errorOut == nil && t.Next != nil {
		if t.Next.ErrorHandler == nil {
			t.Next.ErrorHandler = t.ErrorHandler
		}
		t.ensureBag()
		errorOut = t.Next.execWithBag(t.bag)
	}

	return errorOut
}

// Get the item identified by key, or a default value if the item doesn't exist
func (t *TaskGroup) Get(key string, dflt interface{}) interface{} {
	t.ensureBag()
	val, ok := t.bag.get(key)
	if !ok || val == nil {
		return dflt
	}

	return val
}

// Set the item by key
func (t *TaskGroup) Set(key string, value interface{}) {
	t.ensureBag()
	// don't allow nils in the map
	if value == nil {
		t.bag.remove(key)
	} else {
		t.bag.set(key, value)
	}
}

// Unset an item by key
func (t *TaskGroup) Unset(key string) {
	t.Set(key, nil)
}

func (t *TaskGroup) ensureBag() {
	if t.bag == nil {
		t.bag = &bag{data: make(map[string]interface{})}
	}
}

func (t *TaskGroup) execWithBag(b *bag) error {
	t.ensureBag()
	t.bag.absorb(b)
	return t.Exec()
}
