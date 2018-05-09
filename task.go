package taskgraph

type Task func(*TaskGroup) error
