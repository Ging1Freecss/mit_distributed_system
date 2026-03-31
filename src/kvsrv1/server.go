package kvsrv

import (
	"log"
	"sync"

	"6.5840/kvsrv1/rpc"
	"6.5840/labrpc"
	tester "6.5840/tester1"
)

const Debug = false

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return
}

type key = string
type value struct {
	value   string
	version rpc.Tversion
}

type KVServer struct {
	mu sync.Mutex

	// Your definitions here.
	data map[key]*value
}

func MakeKVServer() *KVServer {
	kv := &KVServer{}
	// Your code here.
	kv.data = make(map[key]*value)
	return kv
}

// Get returns the value and version for args.Key, if args.Key
// exists. Otherwise, Get returns ErrNoKey.
func (kv *KVServer) Get(args *rpc.GetArgs, reply *rpc.GetReply) {
	// Your code here.
	kv.mu.Lock()
	defer kv.mu.Unlock()

	val, ok := kv.data[args.Key]

	if ok == false {
		reply.Err = rpc.ErrNoKey
		reply.Version = 0
		reply.Value = ""
		return
	}

	reply.Err = rpc.OK
	reply.Version = val.version
	reply.Value = val.value
}

// Update the value for a key if args.Version matches the version of
// the key on the server. If versions don't match, return ErrVersion.
// If the key doesn't exist, Put installs the value if the
// args.Version is 0, and returns ErrNoKey otherwise.
func (kv *KVServer) Put(args *rpc.PutArgs, reply *rpc.PutReply) {
	// Your code here.
	kv.mu.Lock()
	defer kv.mu.Unlock()
	
	data, exist := kv.data[args.Key]

	if args.Version == 0 && !exist {
		kv.data[args.Key] = &value{
			value:   args.Value,
			version: 1,
		}

		reply.Err = rpc.OK
	} else if args.Version == 0 && exist {
		reply.Err = rpc.ErrVersion
	} else if args.Version > 0 && !exist {
		reply.Err = rpc.ErrNoKey
	} else if args.Version > 0 && exist && args.Version == data.version {
		kv.data[args.Key].version++
		kv.data[args.Key].value = args.Value
		reply.Err = rpc.OK
	} else if args.Version > 0 && exist && args.Version != data.version {
		reply.Err = rpc.ErrVersion
	}
}

// You can ignore all arguments; they are for replicated KVservers
func StartKVServer(tc *tester.TesterClnt, ends []*labrpc.ClientEnd, gid tester.Tgid, srv int, persister *tester.Persister) []any {
	kv := MakeKVServer()
	return []any{kv}
}
