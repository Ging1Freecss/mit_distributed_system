package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type Coordinator struct {
	// Your definitions here.
	mu sync.Mutex

	// map task
	mapTasks []Task
	nMap     int
	mapsDone int

	reduceTasks []Task
	nReduce     int
	reduceDone  int

	files []string
	done  bool
}

type state_task int

const (
	IDLE state_task = iota
	IN_PROGRESS
	COMPLETED
)

type Task struct {
	status    state_task
	startTime time.Time
}

// Your code here -- RPC handlers for the worker to call.

func (c *Coordinator) GetTask(args *GetTaskArgs, reply *GetTaskReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// check for task that are in progress for too long
	for i := range c.mapTasks {

		t1 := c.mapTasks[i].startTime
		t2 := time.Now()
		diff := t2.Sub(t1)

		if diff > time.Second*10 && c.mapTasks[i].status == IN_PROGRESS {
			c.mapTasks[i].status = state_task(IDLE)
		}
	}

	for i := range c.reduceTasks {

		t1 := c.reduceTasks[i].startTime
		t2 := time.Now()
		diff := t2.Sub(t1)

		if diff > time.Second*10 && c.reduceTasks[i].status == IN_PROGRESS {
			c.reduceTasks[i].status = state_task(IDLE)
		}
	}

	// assign task to map func
	if c.mapsDone < c.nMap {
		for i := range c.mapTasks {
			if c.mapTasks[i].status == state_task(IDLE) {
				c.mapTasks[i].status = state_task(IN_PROGRESS)
				c.mapTasks[i].startTime = time.Now()

				reply.Filename = c.files[i]
				reply.NMap = c.nMap
				reply.NReduce = c.nReduce
				reply.TaskType = task_type(MAP)
				reply.TaskID = i
				return nil
			}
		}
		reply.TaskType = task_type(WAIT)
		return nil
	}

	// now that map is full,
	if c.reduceDone < c.nReduce {
		for i := range c.reduceTasks {
			if c.reduceTasks[i].status == state_task(IDLE) {
				c.reduceTasks[i].status = state_task(IN_PROGRESS)
				c.reduceTasks[i].startTime = time.Now()

				//reply.Filename = c.files[i]
				reply.NMap = c.nMap
				reply.NReduce = c.nReduce
				reply.TaskType = task_type(REDUCE)
				reply.TaskID = i
				return nil
			}
		}
		reply.TaskType = task_type(WAIT)
		return nil
	}
	reply.TaskType = task_type(EXIT)
	return nil
}

func (c *Coordinator) ReportTask(arg *ReportTaskArgs, reply *ReportTaskReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch arg.TaskType {
	case task_type(MAP):
		if c.mapTasks[arg.TaskID].status == state_task(IN_PROGRESS) {
			c.mapTasks[arg.TaskID].status = state_task(COMPLETED)
			c.mapsDone++
		}

	case task_type(REDUCE):
		if c.reduceTasks[arg.TaskID].status == state_task(IN_PROGRESS) {
			c.reduceTasks[arg.TaskID].status = state_task(COMPLETED)
			c.reduceDone++
		}

		if c.reduceDone == c.nReduce {
			c.done = true
		}
	}
	return nil
}

// an example RPC handler.

// the RPC argument and reply types are defined in rpc.go.
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server(sockname string) {
	rpc.Register(c)
	rpc.HandleHTTP()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatalf("listen error %s: %v", sockname, e)
	}
	go http.Serve(l, nil)
}

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Your code here.

	return c.done
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(sockname string, files []string, nReduce int) *Coordinator {
	c := Coordinator{}

	// Your code here.
	var nMap int = len(files)

	// map
	c.nMap = nMap
	c.mapTasks = make([]Task, nMap)
	for i := range c.mapTasks {
		c.mapTasks[i].startTime = time.Now()
		c.mapTasks[i].status = IDLE
	}
	c.mapsDone = 0

	// reduce
	c.reduceTasks = make([]Task, nReduce)
	for i := range c.reduceTasks {
		c.reduceTasks[i].startTime = time.Now()
		c.reduceTasks[i].status = IDLE
	}

	c.nReduce = nReduce
	c.reduceDone = 0

	c.files = files
	c.done = false

	c.server(sockname)
	return &c
}
