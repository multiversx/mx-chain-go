import shutil
from pathlib import Path


def main():
    # Nodes coordinator
    shutil.copyfile("chaosPatching/nodesCoordinator.go.patch", "sharding/nodesCoordinator/chaos.go")

    file_path = Path("sharding/nodesCoordinator/indexHashedNodesCoordinator.go")
    current_content = file_path.read_text()

    marker = "// chaos-testing-point:NewIndexHashedNodesCoordinator_chaosControllerLearnNodes"
    patch = """chaosControllerLearnNodes(currentConfig.eligibleMap, currentConfig.waitingMap)"""
    new_content = current_content.replace(marker, patch)

    marker = "// chaos-testing-point:indexHashedNodesCoordinator_EpochStartPrepare"
    patch = """chaosControllerLearnNodes(resUpdateNodes.Eligible, resUpdateNodes.Waiting)"""
    new_content = new_content.replace(marker, patch)

    file_path.write_text(new_content)

    # Shard processor
    file_path = Path("process/transaction/shardProcess.go")
    current_content = file_path.read_text()
    marker = "// chaos-testing-point:shardProcess_ProcessTransaction"
    patch = """chaos.Controller.CallsCounters.ProcessTransaction.Increment()

	if chaos.Controller.In_shardProcess_processTransaction_shouldReturnError() {
		return vmcommon.ExecutionFailed, chaos.ErrChaoticBehavior
	}
"""
    new_content = current_content.replace(marker, patch)
    new_content = add_chaos_import(new_content)

    file_path.write_text(new_content)


def add_chaos_import(file_content: str) -> str:
    statement = "import \"github.com/multiversx/mx-chain-go/chaos\""
    lines = file_content.split("\n")
    lines.insert(1, statement)
    return "\n".join(lines)


if __name__ == "__main__":
    main()
