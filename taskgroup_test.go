package taskchain

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"reflect"
	"sync"
	"testing"
	"time"
	"strings"
)

var test_mutex sync.RWMutex
var test_bytes []byte
var test_buffer = bytes.NewBuffer(test_bytes)
var test_logger = log.New(test_buffer, "", 0)

func passTask(t *TaskGroup) error {
	fmt.Println("this task will succeed")
	return nil
}

func failTask(t *TaskGroup) error {
	fmt.Println("this task will fail")
	return errors.New("fail")
}

func bagAddTask(t *TaskGroup) error {
	test_mutex.Lock()
	a := t.Get("football", 0).(int)
	a++
	t.Set("football", a)
	test_mutex.Unlock()
	return nil
}

func errHandler(t *TaskGroup, err error) {
	test_logger.Print("*")
}

func TestTaskGroup_Add(t *testing.T) {
	type args struct {
		task Task
	}
	tasksMatch := func(t1, t2 []Task) bool {
		if len(t1) != len(t2) {
			return false
		}
		for i := range t1 {
			v1 := reflect.ValueOf(t1[i]).Pointer()
			v2 := reflect.ValueOf(t2[i]).Pointer()
			if v1 != v2 {
				return false
			}
		}

		return true
	}
	tests := []struct {
		name     string
		t        *TaskGroup
		args     args
		expected []Task
	}{
		{
			name: "add one task",
			t:    &TaskGroup{},
			args: args{
				task: passTask,
			},
			expected: []Task{passTask},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.t.Add(tt.args.task)
			if !tasksMatch(tt.t.tasks, tt.expected) {
				t.Errorf("Task.Add() got = %v, want %v", tt.t.tasks, tt.expected)
			}
		})
	}
}

func TestTaskGroup_Exec(t *testing.T) {
	tests := []struct {
		name    string
		t       *TaskGroup
		wantErr bool
	}{
		{
			name: "simple pass",
			t: &TaskGroup{
				tasks: []Task{passTask},
			},
			wantErr: false,
		},
		{
			name: "simple fail",
			t: &TaskGroup{
				tasks: []Task{failTask},
			},
			wantErr: true,
		},
		{
			name: "compound pass",
			t: &TaskGroup{
				tasks: []Task{passTask, passTask},
			},
			wantErr: false,
		},
		{
			name: "compound fail",
			t: &TaskGroup{
				tasks: []Task{failTask, failTask},
			},
			wantErr: true,
		},
		{
			name: "mixed pass/fail",
			t: &TaskGroup{
				tasks: []Task{passTask, failTask},
			},
			wantErr: true,
		},
		{
			name: "many pass",
			t: &TaskGroup{
				tasks: []Task{
					passTask, passTask, passTask, passTask, passTask, passTask,
					passTask, passTask, passTask, passTask, passTask, passTask,
					passTask, passTask, passTask, passTask, passTask, passTask,
					passTask, passTask, passTask, passTask, passTask, passTask,
					passTask, passTask, passTask, passTask, passTask, passTask},
			},
			wantErr: false,
		},
		{
			name: "turtles pass",
			t: &TaskGroup{
				tasks: []Task{passTask},
				Next: &TaskGroup{
					tasks: []Task{passTask},
					Next: &TaskGroup{
						tasks: []Task{passTask},
						Next: &TaskGroup{
							tasks: []Task{passTask},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "turtles fail",
			t: &TaskGroup{
				tasks: []Task{passTask},
				Next: &TaskGroup{
					tasks: []Task{passTask},
					Next: &TaskGroup{
						tasks: []Task{passTask},
						Next: &TaskGroup{
							tasks: []Task{failTask},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.t.Exec(); (err != nil) != tt.wantErr {
				t.Errorf("TaskGroup.Exec() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTaskGroup_Football(t *testing.T) {
	l3 := &TaskGroup{
		tasks: []Task{bagAddTask, bagAddTask, bagAddTask, bagAddTask, bagAddTask},
	}
	l2 := &TaskGroup{
		tasks: []Task{bagAddTask, bagAddTask, bagAddTask, bagAddTask},
		Next:  l3,
	}
	l1 := &TaskGroup{
		tasks: []Task{bagAddTask, bagAddTask, bagAddTask},
		Next:  l2,
	}
	t1 := &TaskGroup{
		tasks: []Task{bagAddTask, bagAddTask},
		Next:  l1,
	}

	err := t1.Exec()
	if err != nil {
		t.Fail()
	}

	v := l3.Get("football", 0).(int)
	if v != 14 {
		t.Fail()
	}
}

func TestTaskGroup_Get(t *testing.T) {
	type args struct {
		key  string
		dflt interface{}
	}
	tests := []struct {
		name string
		t    *TaskGroup
		args args
		want interface{}
	}{
		{
			name: "simple get",
			t: &TaskGroup{
				bag: &bag{
					data: map[string]interface{}{
						"foo": "bar",
					},
				},
			},
			args: args{
				key:  "foo",
				dflt: "",
			},
			want: "bar",
		},
		{
			name: "empty get",
			t: &TaskGroup{
				bag: &bag{
					data: make(map[string]interface{}),
				},
			},
			args: args{
				key:  "foo",
				dflt: "oof",
			},
			want: "oof",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t.Get(tt.args.key, tt.args.dflt); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TaskGroup.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskGroup_Set(t *testing.T) {
	type args struct {
		key   string
		value interface{}
	}
	tests := []struct {
		name     string
		t        *TaskGroup
		args     args
		expected map[string]interface{}
	}{
		{
			name: "simple set",
			t:    &TaskGroup{},
			args: args{
				key:   "foo",
				value: "bar",
			},
			expected: map[string]interface{}{
				"foo": "bar",
			},
		},
		{
			name: "nil set",
			t: &TaskGroup{
				bag: &bag{
					data: map[string]interface{}{
						"foo":  "bar",
						"fizz": "buzz",
					},
				},
			},
			args: args{
				key:   "foo",
				value: nil,
			},
			expected: map[string]interface{}{
				"fizz": "buzz",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.t.Set(tt.args.key, tt.args.value)
			if !reflect.DeepEqual(tt.t.bag.data, tt.expected) {
				t.Errorf("bag.absorb() got = %v, want %v", tt.t.bag.data, tt.expected)
			}
		})
	}
}

func TestTaskGroup_Unset(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name     string
		t        *TaskGroup
		args     args
		expected map[string]interface{}
	}{
		{
			name: "simple unset",
			t: &TaskGroup{
				bag: &bag{
					data: map[string]interface{}{
						"foo":  "bar",
						"fizz": "buzz",
					},
				},
			},
			args: args{
				key: "foo",
			},
			expected: map[string]interface{}{
				"fizz": "buzz",
			},
		},
		{
			name: "idempotent unset",
			t: &TaskGroup{
				bag: &bag{
					data: map[string]interface{}{
						"fizz": "buzz",
					},
				},
			},
			args: args{
				key: "foo",
			},
			expected: map[string]interface{}{
				"fizz": "buzz",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.t.Unset(tt.args.key)
		})
	}
}

func TestTaskGroup_ensureBag(t *testing.T) {
	tests := []struct {
		name string
		t    *TaskGroup
	}{
		{
			name: "simple ensure",
			t:    &TaskGroup{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.t.bag != nil {
				t.Fail()
			}
			tt.t.ensureBag()
			if tt.t.bag == nil {
				t.Fail()
			}
		})
	}
}

func TestTaskGroup_errHandler(t *testing.T) {
	failMsg := strings.TrimSpace(`
*
*
*
*
*
`)
	t1 := &TaskGroup{
		tasks: []Task{failTask, failTask, failTask, failTask, failTask},
	}
	t1.ErrorHandler = errHandler

	t1.Exec()
	time.Sleep(10 * time.Millisecond)
	buffout := strings.TrimSpace(test_buffer.String())
	fmt.Println(buffout)
	if buffout != failMsg {
		t.Fail()
	}
}
