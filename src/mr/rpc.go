package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.
type GetTaskArgs struct{}

type task_type int

const (
	MAP task_type = iota
	REDUCE
	WAIT
	EXIT
)

type GetTaskReply struct {
	TaskType task_type
	TaskID   int
	Filename string
	NReduce  int
	NMap     int
}

type ReportTaskArgs struct{
	TaskType task_type
	TaskID int
}

type ReportTaskReply struct{}