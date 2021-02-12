`batcher` is a generic library to process objects in batches. A producer adds
objects into the batcher, and the batcher asynchronously triggers the processor
function with a batch of objects under 1 of 2 conditions:

    1. The max buffered items has been reached
    1. The flush period has occurred
