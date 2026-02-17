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