import shutil
from pathlib import Path
from typing import Tuple


def main():
    shutil.copyfile("chaosPatching/nodesCoordinator.go.patch", "sharding/nodesCoordinator/chaos.go")

    do_replacements(
        file_path=Path("sharding/nodesCoordinator/indexHashedNodesCoordinator.go"),
        replacements=[
            (
                "// chaos-testing-point:NewIndexHashedNodesCoordinator_chaosControllerLearnNodes",
                """chaosControllerLearnNodes(currentConfig.eligibleMap, currentConfig.waitingMap)""",
            ),
            (
                "// chaos-testing-point:indexHashedNodesCoordinator_EpochStartPrepare",
                """chaosControllerLearnNodes(resUpdateNodes.Eligible, resUpdateNodes.Waiting)"""
            )
        ]
    )

    do_replacements(
        file_path=Path("process/transaction/shardProcess.go"),
        replacements=[
            (
                "// chaos-testing-point:shardProcess_ProcessTransaction",
                """chaos.Controller.CallsCounters.ProcessTransaction.Increment()

    if chaos.Controller.In_shardProcess_processTransaction_shouldReturnError() {
        return vmcommon.ExecutionFailed, chaos.ErrChaoticBehavior
	}
"""
            ),
        ],
        with_import=True
    )

    do_replacements(
        file_path=Path("consensus/spos/bls/v1/subroundSignature.go"),
        replacements=[
            (
                "// chaos-testing-point:v1/subroundSignature_doSignatureJob_corruptSignatureWhenSingleKey",
                """chaos.Controller.In_subroundSignature_doSignatureJob_maybeCorruptSignature_whenSingleKey(sr.GetHeader(), signatureShare)"""
            ),
            (
                "// chaos-testing-point:v1/subroundSignature_doSignatureJob_corruptSignatureWhenMultiKey",
                """chaos.Controller.In_subroundSignature_doSignatureJob_maybeCorruptSignature_whenMultiKey(sr.GetHeader(), selfIndex, signatureShare)"""
            ),
            (
                "// chaos-testing-point:v1/subroundSignature_completeSignatureSubRound_skipWaitingForSignatures",
                """if chaos.Controller.In_subroundSignature_completeSignatureSubRound_shouldSkipWaitingForSignatures(sr.GetHeader()) {
		return true
	}
"""
            )
        ],
        with_import=True
    )

    do_replacements(
        file_path=Path("consensus/spos/bls/v1/subroundEndRound.go"),
        replacements=[
            (
                "// chaos-testing-point:v1/subroundEndRound_checkSignaturesValidity_returnError",
                """if chaos.Controller.In_subroundEndRound_checkSignaturesValidity_shouldReturnError(sr.GetHeader()) {
		return spos.ErrInvalidSignature
	}
"""
            ),
        ],
        with_import=True
    )

    do_replacements(
        file_path=Path("consensus/spos/bls/v2/subroundBlock.go"),
        replacements=[
            (
                "// chaos-testing-point:v2/subroundBlock_doBlockJob_corruptLeaderSignature",
                """chaos.Controller.In_V2_subroundBlock_doBlockJob_maybeCorruptLeaderSignature(header, leaderSignature)"""
            ),
            (
                "// chaos-testing-point:v2/subroundBlock_doBlockJob_skipSendingBlock",
                """if chaos.Controller.In_V2_subroundBlock_doBlockJob_shouldSkipSendingBlock(header) {
        return false
	}
"""
            )
        ],
        with_import=True
    )


def do_replacements(file_path: Path, replacements: list[Tuple[str, str]], with_import: bool = False):
    content = file_path.read_text()

    for marker, patch in replacements:
        content = content.replace(marker, patch)

    if with_import:
        content = add_chaos_import(content)

    file_path.write_text(content)


def add_chaos_import(file_content: str) -> str:
    statement = "import \"github.com/multiversx/mx-chain-go/chaos\""
    lines = file_content.split("\n")
    lines.insert(1, statement)
    return "\n".join(lines)


if __name__ == "__main__":
    main()
