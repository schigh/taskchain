package taskchain

// Task is any function that takes in a TaskGroup and returns an error
type Task func(*TaskGroup) error
