# MIT 6.5840 - Distributed Systems Labs (Spring 2025)

My solutions for the [MIT 6.5840 Distributed Systems](https://pdos.csail.mit.edu/6.824/) course labs.

## Lab 1: MapReduce

A distributed MapReduce implementation with a coordinator and multiple workers communicating via RPC.

### Architecture
- **Coordinator** (`src/mr/coordinator.go`) — assigns Map/Reduce tasks, handles 10s worker timeouts
- **Worker** (`src/mr/worker.go`) — executes Map and Reduce tasks in a loop
- **RPC** (`src/mr/rpc.go`) — defines message types for coordinator-worker communication

### Running Tests

```bash
cd src
make mr                        # run all MapReduce tests
make RUN="-run Wc" mr          # run only the word count test
```

### Test Results
- ✅ TestWc — word count correctness
- ✅ TestIndexer — inverted index correctness
- ✅ TestMapParallel — mappers run in parallel
- ✅ TestReduceParallel — reducers run in parallel
- ✅ TestJobCount — correct number of map jobs
- ✅ TestEarlyExit — no premature exit
- ✅ TestCrashWorker — crash recovery with 10s timeout

## Lab 2: Key/Value Server

A linearizable key/value server with version-based conditional puts, at-most-once semantics on unreliable networks, and a distributed lock built on top.

### Architecture
- **Server** (`src/kvsrv1/server.go`) — in-memory KV store with `sync.Mutex`, supports `Get(key)` and `Put(key, value, version)` with version-based conditional writes
- **Client** (`src/kvsrv1/client.go`) — `Clerk` that sends RPCs, retries on network failure, returns `ErrMaybe` when a retransmitted Put gets `ErrVersion` (uncertain if original succeeded)
- **Lock** (`src/kvsrv1/lock/lock.go`) — distributed lock using CAS via `Put` version checks; uses a unique `clientID` to verify ownership after `ErrMaybe`

### Key Concepts
- **Version numbers** — each key has a version that increments on successful Put, enabling optimistic concurrency control (compare-and-swap)
- **At-most-once semantics** — retried Puts are naturally deduplicated by the version check; `ErrMaybe` signals uncertainty to the application
- **Distributed lock** — Acquire spins with Get+Put until the lock key's value equals the client's ID; Release sets the value to empty

### Running Tests

```bash
cd src
make kvsrv1                          # run all KV server tests
make RUN="-run Reliable" kvsrv1      # reliable network only
make lock1                           # run all lock tests
make RUN="-run Reliable" lock1       # reliable network only
```

### Test Results — KV Server
- ✅ TestReliablePut — single client Put/Get correctness
- ✅ TestPutConcurrentReliable — 10 clients racing on same key
- ✅ TestMemPutManyClientsReliable — 20K clients, no memory leak
- ✅ TestUnreliableNet — single client with dropped messages, ErrMaybe handling

### Test Results — Lock
- ✅ TestReliableBasic — single Acquire and Release
- ✅ TestReliableNested — two independent locks held concurrently
- ✅ TestOneClientReliable — 1 client, repeated lock cycles
- ✅ TestManyClientsReliable — 10 clients competing for locks
- ✅ TestOneClientUnreliable — 1 client with unreliable network
- ✅ TestManyClientsUnreliable — 10 clients with unreliable network