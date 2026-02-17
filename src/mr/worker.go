package mr

import (
	"cmp"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/rpc"
	"os"
	"slices"
	"time"
)

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

var coordSockName string // socket for coordinator

// main/mrworker.go calls this function.
func Worker(sockname string, mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	coordSockName = sockname

	// Your worker implementation here.
	for {
		reply := &GetTaskReply{}
		arg := &GetTaskArgs{}
		state := call("Coordinator.GetTask", arg, reply)

		if !state || reply.TaskType == task_type(EXIT) {
			return
		}

		switch reply.TaskType {

		case task_type(MAP):
			doMapTask(reply, mapf)

			reportArg := &ReportTaskArgs{TaskType: reply.TaskType, TaskID: reply.TaskID}
			reportReply := &ReportTaskReply{}

			call("Coordinator.ReportTask", reportArg, reportReply)

		case task_type(REDUCE):
			doReduceTask(reply, reducef)

			reportArg := &ReportTaskArgs{TaskType: reply.TaskType, TaskID: reply.TaskID}
			reportReply := &ReportTaskReply{}

			call("Coordinator.ReportTask", reportArg, reportReply)

		case task_type(WAIT):
			time.Sleep(time.Second)
			continue
		}

	}

}

// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.Example", &args, &reply)
	if ok {
		// reply.Y should be 100.
		fmt.Printf("reply.Y %v\n", reply.Y)
	} else {
		fmt.Printf("call failed!\n")
	}
}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	c, err := rpc.DialHTTP("unix", coordSockName)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	if err := c.Call(rpcname, args, reply); err == nil {
		return true
	}
	log.Printf("%d: call failed err %v", os.Getpid(), err)
	return false
}

func doMapTask(reply *GetTaskReply, mapf func(string, string) []KeyValue) {

	data, err := os.ReadFile(reply.Filename)

	if err != nil {
		return
	}

	key_value := mapf(reply.Filename, string(data))
	bucket := make([][]KeyValue, reply.NReduce)

	for _, val := range key_value {
		i := ihash(val.Key) % reply.NReduce
		bucket[i] = append(bucket[i], val)
	}

	for i := 0; i < reply.NReduce; i++ {
		tmpFile, err := os.CreateTemp("", "mr-temp-*")

		if err != nil {
			return
		}

		encoder := json.NewEncoder(tmpFile)
		for _, v := range bucket[i] {
			encoder.Encode(&v)
		}

		tmpFile.Close()
		os.Rename(tmpFile.Name(), fmt.Sprintf("mr-%d-%d", reply.TaskID, i))
	}
}

func doReduceTask(reply *GetTaskReply, reducef func(string, []string) string) {

	kva := []KeyValue{}

	for i := 0; i < reply.NMap; i++ {
		fileName := fmt.Sprintf("mr-%v-%v", i, reply.TaskID)

		file, err := os.Open(fileName)

		if err != nil {
			return
		}
		decoder := json.NewDecoder(file)

		for decoder.More() {
			var k KeyValue
			err := decoder.Decode(&k)

			if err != nil {
				if err == io.EOF {
					break
				}
				log.Fatal(err)
			}
			kva = append(kva, k)
		}

		file.Close()
	}

	slices.SortFunc(kva, func(a, b KeyValue) int {
		return cmp.Compare(a.Key, b.Key)
	})

	tmpFile, err := os.CreateTemp("", "mr-out-tmp-*")

	if err != nil {
		return
	}

	i := 0
	for i < len(kva) {

		j := i + 1
		for j < len(kva) && kva[j].Key == kva[i].Key {
			j++
		}

		temp_kva := []string{}

		for idx := i; idx < j; idx++ {
			temp_kva = append(temp_kva, kva[idx].Value)
		}

		output := reducef(kva[i].Key, temp_kva)

		tmpFile.WriteString(fmt.Sprintf("%v %v\n", kva[i].Key, output))
		i = j
	}
	tmpFile.Close()
	os.Rename(tmpFile.Name(), fmt.Sprintf("mr-out-%v", reply.TaskID))

}
