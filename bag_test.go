package taskchain

import (
	"reflect"
	"testing"
)

func Test_bag_get(t *testing.T) {
	type args struct {
		key string
	}

	bag1 := &bag{
		data: map[string]interface{}{
			"foo": "bar",
		},
	}

	tests := []struct {
		name  string
		b     *bag
		args  args
		want  interface{}
		want1 bool
	}{
		{
			name: "happy path",
			b:    bag1,
			args: args{
				key: "foo",
			},
			want:  "bar",
			want1: true,
		},
		{
			name: "happy path 2",
			b:    bag1,
			args: args{
				key: "fooz",
			},
			want:  nil,
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.b.get(tt.args.key)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("bag.get() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("bag.get() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_bag_set(t *testing.T) {
	type args struct {
		key   string
		value interface{}
	}

	testBag1 := &bag{
		data: map[string]interface{}{
			"fizz": "buzz",
		},
	}

	testBag2 := &bag{
		data: map[string]interface{}{
			"foo": "bazz",
		},
	}
	tests := []struct {
		name     string
		b        *bag
		args     args
		expected map[string]interface{}
	}{
		{
			name: "empty bag",
			b:    &bag{},
			args: args{
				key:   "foo",
				value: "bar",
			},
			expected: map[string]interface{}{
				"foo": "bar",
			},
		},
		{
			name: "bag with items",
			b:    testBag1,
			args: args{
				key:   "foo",
				value: "bar",
			},
			expected: map[string]interface{}{
				"foo":  "bar",
				"fizz": "buzz",
			},
		},
		{
			name: "overwritten item",
			b:    testBag2,
			args: args{
				key:   "foo",
				value: "bar",
			},
			expected: map[string]interface{}{
				"foo": "bar",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.b.set(tt.args.key, tt.args.value)
			if !reflect.DeepEqual(tt.b.data, tt.expected) {
				t.Errorf("bag.get() got = %v, want %v", tt.b.data, tt.expected)
			}
		})
	}
}

func Test_bag_remove(t *testing.T) {
	type args struct {
		key string
	}

	testBag1 := &bag{
		data: map[string]interface{}{
			"foo":  "bar",
			"fizz": "buzz",
		},
	}

	testBag2 := &bag{
		data: map[string]interface{}{
			"fizz": "buzz",
		},
	}

	testBag3 := &bag{}

	tests := []struct {
		name     string
		b        *bag
		args     args
		expected map[string]interface{}
	}{
		{
			name: "remove item",
			b:    testBag1,
			args: args{
				"foo",
			},
			expected: map[string]interface{}{
				"fizz": "buzz",
			},
		},
		{
			name: "empty bag",
			b:    &bag{},
			args: args{
				"foo",
			},
			expected: testBag3.data,
		},
		{
			name: "nonexistent key",
			b:    testBag2,
			args: args{
				"foo",
			},
			expected: map[string]interface{}{
				"fizz": "buzz",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.b.remove(tt.args.key)
			if !reflect.DeepEqual(tt.b.data, tt.expected) {
				t.Errorf("bag.remove() got = %v, want %v", tt.b.data, tt.expected)
			}
		})
	}
}

func Test_bag_absorb(t *testing.T) {
	type args struct {
		other *bag
	}

	testBag1 := &bag{
		data: map[string]interface{}{
			"foo":  "bar",
			"fizz": "buzz",
		},
	}

	testBag2 := &bag{
		data: map[string]interface{}{
			"foo":  "bazz",
			"herp": "derp",
		},
	}

	tests := []struct {
		name     string
		b        *bag
		args     args
		expected map[string]interface{}
	}{
		{
			name: "absorb empty bag",
			b:    testBag1,
			args: args{
				other: &bag{},
			},
			expected: testBag1.data,
		},
		{
			name: "absorb other bag",
			b:    testBag1,
			args: args{
				other: testBag2,
			},
			expected: map[string]interface{}{
				"foo":  "bar",
				"fizz": "buzz",
				"herp": "derp",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.b.absorb(tt.args.other)
			if !reflect.DeepEqual(tt.b.data, tt.expected) {
				t.Errorf("bag.absorb() got = %v, want %v", tt.b.data, tt.expected)
			}
		})
	}
}
