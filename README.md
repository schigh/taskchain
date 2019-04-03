# Deprecated
Don't use this.  Use `golang.org/x/sync/errgroup` instead.

----

[![LICENSE](https://img.shields.io/badge/license-MIT-orange.svg)](LICENSE)
[![Build Status](https://travis-ci.org/schigh/taskchain.svg?branch=master)](https://travis-ci.org/schigh/taskchain)
[![codecov](https://codecov.io/gh/schigh/taskchain/branch/master/graph/badge.svg?token=hhqA1l88kx)](https://codecov.io/gh/schigh/taskchain)
[![Go Report Card](https://goreportcard.com/badge/github.com/schigh/taskchain)](https://goreportcard.com/report/github.com/schigh/taskchain)
[![Godocs](https://img.shields.io/badge/golang-documentation-blue.svg)](https://godoc.org/github.com/schigh/taskchain)

# taskchain

> Simple barrier logic for asynchronous non-cancellable tasks.

taskchain is a very simple implementation of what's known as a [barrier](https://en.wikipedia.org/wiki/Barrier_(computer_science)).  TL;DR: a barrier is a mechanism that aggregates a group of asynchronous tasks in such a way that they all must complete before the next barrier can begin.

The task chain consists of one to many contiguous task groups.  Within each task group is one to many tasks.  A task is a function with this signature:

```go
type Task func(*TaskGroup) error
```

We'll see why you would want to take in the task group as a parameter in a moment.  What's important to know here is that _if your task returns an error, the task group will not continue to the next task group_.  More specifically, task group is a struct:

```go
type TaskGroup struct {
    Next *TaskGroup
    ErrorHandler func(*TaskGroup, error)
}
```

You add tasks to a task group with the `Add` function:

```go
func (t *TaskGroup) Add(task Task) {...}
```

```go
func myTask (t *TaskGroup) error {
    // do stuff
    return nil
}

tg := &TaskGroup{}
tg.Add(myTask)
```



`TaskGroup` tasks are dispatched with the `Exec` function:

```go
func (t *TaskGroup) Exec() error {...}
```

```go
tg := &TaskGroup{}
tg.Add(func(t *TaskGroup) error {
    // do stuff
    return nil
})

if err := tg.Exec(); err != nil {
    // handle error
}
```

If _any_ of the tasks within the task group fail, then `Exec` returns **the first** error returned by any of the tasks.  However, you can listen for any errors that occur when the task group executes by setting its error handler:

```go
func errHandler(t *TaskGroup, err error) {
    // this error handler will be called for ANY 
    // errors returned while the task group is executing
}

tg := &TaskGroup{}
tg.ErrorHandler = errHandler
```

You can inject dependencies into a task group by accessing its _bag_.  The bag is an internal map that is mutexed during storage and retrieval operations.  You are responsible for the goroutine safety of whatever you put in the bag, however.

You add things to the bag like so:

```go
tg := &TaskGroup{}
tg.Set("foo", "bar")
```

And then access it within your task like so:

```go
func myTask(t *TaskGroup) error {
    fooThing := t.Get("foo", "bazz").(string)
}
```

Note that second argument in `t.Get`.  If there is no item in the bag with key `foo`, it will return a default value (`bazz` in this case).  If there is no item matching the key, or if the item matching the key is `nil`, then the default will be used (this is a straight ripoff of Python's `get` dictionary function).  If you want `nil` to be returned if the key is not found, set that second parameter to `nil`.

Also note that `Get` returns an `interface{}`, so you need to cast it to the appropriate type.

One nice thing about the bag is that it gets passed from parent task group to child task group.  Well, it's not so much passed along as it is overlayed atop the child group's bag.  If a child task group has an item in its bag with the same key as its parent, the child's value will _not_ be overwritten.  For example:

```go
tgParent := &TaskGroup{}
tgParent.Set("foo", "bar")
tgParent.Set("fizz", "buzz")
// add tasks to parent ...
tgChild := &TaskGroup{}
tgChild.Set("foo", "bazz")
// add tasks to child ...

tgParent.Next = tgChild
_ = tgParent.Exec()

// this will be 'bazz', not 'bar'
v1 := tgChild.Get("foo", "").(string)

// this will be 'buzz' inherited from the parent task group
v2 := tgChild.Get("fizz", "").(string)
```

The same forwarding logic applies to error handlers.  If the parent task group has an error handler and the child task group does not, then the child task group will inherit the parent's handler.  If the child group has its own error handler, that will be used.

