package constants

const (
	BlockPending   = 0
	BlockProcessed = 1
	BlockFailed    = 2
	// BlockSkipped GetBlock err:{"code":-32007,"message":"Slot 311350484 was skipped, or missing due to ledger jump to recent snapshot","data":null}
	BlockSkipped = 3
)

// SPL Token instruction discriminators
const (
	InstructionTransfer        = 3
	InstructionTransferChecked = 12
)
