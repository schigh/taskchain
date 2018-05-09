package taskgraph

import "sync"

type TaskGroup struct {
	tasks    []Task
	bag      *bag
	Policies GroupPolicy
	Next     *TaskGroup
}

func (t *TaskGroup) Add(task Task) {
	t.tasks = append(t.tasks, task)
}

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
			if containsPolicy(t.Policies, HaltOnAnyError) {
				firstErrorToken.Do(func() {
					errorOut = err
				})
			}
		}
	}

	if errorOut == nil && t.Next != nil {
		errorOut = t.Next.execWithBag(t.bag)
	}

	return errorOut
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

func (t *TaskGroup) Get(key string) interface{} {
	return t.bag.get(key)
}

func (t *TaskGroup) Set(key string, value interface{}) {
	t.ensureBag()
	// don't allow nils in the map
	if value == nil {
		t.bag.remove(key)
	} else {
		t.bag.set(key, value)
	}
}
