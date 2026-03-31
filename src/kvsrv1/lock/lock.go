package lock

import (
	"6.5840/kvsrv1/rpc"
	kvtest "6.5840/kvtest1"
)

type Lock struct {
	// IKVClerk is a go interface for k/v clerks: the interface hides
	// the specific Clerk type of ck but promises that ck supports
	// Put and Get.  The tester passes the clerk in when calling
	// MakeLock().
	ck kvtest.IKVClerk

	// You may add code here
	lockname string
	clientID string
}

// The tester calls MakeLock() and passes in a k/v clerk; your code can
// perform a Put or Get by calling lk.ck.Put() or lk.ck.Get().
//
// This interface supports multiple locks by means of the
// lockname argument; locks with different names should be
// independent.
func MakeLock(ck kvtest.IKVClerk, lockname string) *Lock {
	lk := &Lock{ck: ck,
		lockname: lockname,
		clientID: kvtest.RandValue(64)}
	// You may add code here
	return lk
}

func (lk *Lock) Acquire() {
	// Your code here
	for {
		val, version, err := lk.ck.Get(lk.lockname)

		if err == rpc.ErrNoKey { // create a new lock

			if lk.ck.Put(lk.lockname, lk.clientID, 0) == rpc.OK {
				return
			}

		} else if val == "" { // lock is free

			if lk.ck.Put(lk.lockname, lk.clientID, version) == rpc.OK {
				return
			}
		} else if val == lk.clientID { // lock is already acquire , err maybe case
			return
		}
	}

}

func (lk *Lock) Release() {
	// Your code here

	for {
		_, version, _ := lk.ck.Get(lk.lockname)
		if lk.ck.Put(lk.lockname, "", version) == rpc.OK {
			return
		}
	}

}
