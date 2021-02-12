package batcher

// User-supplied function to process a batch of data. Batcher implementations
// should guarantee that multiple invocations of the processor is not done
// concurrently. The user should handle synchronization if the processing is
// asynchronous.
type BatchProcessor func(batch []interface{})
