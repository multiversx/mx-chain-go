package txcache

const diagnosisMaxTransactionsToDisplay = 10000
const initialCapacityOfSelectionSlice = 30000
const selectionLoopDurationCheckInterval = 10

// maxNumBytesPerSenderUpperBoundTest is used for setting the MaxNumBytesPerSenderUpperBoundTest from ConfigSourceMe in tests
const maxNumBytesPerSenderUpperBoundTest = 33_554_432 // 32 MB
