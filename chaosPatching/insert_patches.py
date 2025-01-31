import shutil
from pathlib import Path


def main():
    # Nodes coordinator
    shutil.copyfile("chaosPatching/nodesCoordinator.go.patch", "sharding/nodesCoordinator/chaos.go")

    file_path = Path("sharding/nodesCoordinator/indexHashedNodesCoordinator.go")
    content = file_path.read_text()

    marker = "// chaos-testing-point:NewIndexHashedNodesCoordinator_chaosControllerLearnNodes"
    patch = """chaosControllerLearnNodes(currentConfig.eligibleMap, currentConfig.waitingMap)"""
    content = content.replace(marker, patch)

    marker = "// chaos-testing-point:indexHashedNodesCoordinator_EpochStartPrepare"
    patch = """chaosControllerLearnNodes(resUpdateNodes.Eligible, resUpdateNodes.Waiting)"""
    content = content.replace(marker, patch)

    file_path.write_text(content)

    # Shard processor
    file_path = Path("process/transaction/shardProcess.go")
    content = file_path.read_text()
    marker = "// chaos-testing-point:shardProcess_ProcessTransaction"
    patch = """chaos.Controller.CallsCounters.ProcessTransaction.Increment()

	if chaos.Controller.In_shardProcess_processTransaction_shouldReturnError() {
		return vmcommon.ExecutionFailed, chaos.ErrChaoticBehavior
	}
"""
    content = content.replace(marker, patch)

    content = add_chaos_import(content)
    file_path.write_text(content)

    # Consensus V2, subround BLOCK
    file_path = Path("consensus/spos/bls/v2/subroundBlock.go")
    content = file_path.read_text()
    marker = "// chaos-testing-point:v2/subroundBlock_doBlockJob_corruptLeaderSignature"
    patch = """chaos.Controller.In_V2_subroundBlock_doBlockJob_maybeCorruptLeaderSignature(header, leaderSignature)"""
    content = content.replace(marker, patch)

    marker = "// chaos-testing-point:v2/subroundBlock_doBlockJob_skipSendingBlock"
    patch = """if chaos.Controller.In_V2_subroundBlock_doBlockJob_shouldSkipSendingBlock(header) {
		return false
	}
    """
    content = content.replace(marker, patch)

    content = add_chaos_import(content)
    file_path.write_text(content)


def add_chaos_import(file_content: str) -> str:
    statement = "import \"github.com/multiversx/mx-chain-go/chaos\""
    lines = file_content.split("\n")
    lines.insert(1, statement)
    return "\n".join(lines)


if __name__ == "__main__":
    main()
