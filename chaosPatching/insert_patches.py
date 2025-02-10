from pathlib import Path
from typing import Tuple


def main():
    do_replacements(
        file_path=Path("cmd/node/main.go"),
        replacements=[
            (
                "// chaos-testing-point/node_main_learnNodeDisplayName",
                """chaos.Controller.LearnNodeDisplayName(preferencesConfig.Preferences.NodeDisplayName)"""
            )
        ],
        with_import=True
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
                """chaos.Controller.In_subroundSignature_doSignatureJob_maybeCorruptSignature_whenSingleKey(sr.GetHeader(), selfIndex, signatureShare)"""
            ),
            (
                "// chaos-testing-point:v1/subroundSignature_doSignatureJob_corruptSignatureWhenMultiKey",
                """chaos.Controller.In_subroundSignature_doSignatureJob_maybeCorruptSignature_whenMultiKey(sr.GetHeader(), selfIndex, signatureShare)"""
            ),
            (
                "// chaos-testing-point:v1/subroundSignature_completeSignatureSubRound_skipWaitingForSignatures",
                """selfIndex, err := sr.SelfConsensusGroupIndex()
    if err != nil {
        selfIndex = -1
    }

    if chaos.Controller.In_subroundSignature_completeSignatureSubRound_shouldSkipWaitingForSignatures(sr.GetHeader(), selfIndex) {
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
                """selfIndex, err := sr.SelfConsensusGroupIndex()
    if err != nil {
        selfIndex = -1
    }
    
    if chaos.Controller.In_subroundEndRound_checkSignaturesValidity_shouldReturnError(sr.GetHeader(), selfIndex) {
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
                """selfIndex, err := sr.SelfConsensusGroupIndex()
    if err != nil {
        selfIndex = -1
    }
    
    chaos.Controller.In_V2_subroundBlock_doBlockJob_maybeCorruptLeaderSignature(header, selfIndex, leaderSignature)"""
            ),
            (
                "// chaos-testing-point:v2/subroundBlock_doBlockJob_skipSendingBlock",
                """selfIndex, errChaos := sr.SelfConsensusGroupIndex()
    if errChaos != nil {
        selfIndex = -1
    }
    
    if chaos.Controller.In_V2_subroundBlock_doBlockJob_shouldSkipSendingBlock(header, selfIndex) {
        return false
	}
"""
            )
        ],
        with_import=True
    )

    do_replacements(
        file_path=Path("consensus/spos/bls/v2/subroundSignature.go"),
        replacements=[
            (
                "// chaos-testing-point:v2/subroundSignature_doSignatureJob_corruptSignatureWhenSingleKey",
                """chaos.Controller.In_subroundSignature_doSignatureJob_maybeCorruptSignature_whenSingleKey(sr.GetHeader(), selfIndex, signatureShare)"""
            ),
            (
                "// chaos-testing-point:v2/subroundSignature_doSignatureJob_corruptSignatureWhenMultiKey",
                """chaos.Controller.In_subroundSignature_doSignatureJob_maybeCorruptSignature_whenMultiKey(sr.GetHeader(), idx, signatureShare)"""
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
