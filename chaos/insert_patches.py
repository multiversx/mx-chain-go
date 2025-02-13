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
                """if chaos.Controller.In_shardProcess_processTransaction_shouldReturnError() {
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
                """chaos.Controller.In_V1_and_V2_subroundSignature_doSignatureJob_maybeCorruptSignature_whenSingleKey(sr, signatureShare)"""
            ),
            (
                "// chaos-testing-point:v1/subroundSignature_doSignatureJob_corruptSignatureWhenMultiKey",
                """chaos.Controller.In_V1_and_V2_subroundSignature_doSignatureJob_maybeCorruptSignature_whenMultiKey(sr, pk, signatureShare)"""
            ),
            (
                "// chaos-testing-point:v1/subroundSignature_completeSignatureSubRound_skipWaitingForSignatures",
                """if chaos.Controller.In_V1_subroundSignature_completeSignatureSubRound_shouldSkipWaitingForSignatures(sr) {
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
                """if chaos.Controller.In_V1_subroundEndRound_checkSignaturesValidity_shouldReturnError(sr) {
		return spos.ErrInvalidSignature
	}
"""
            ),
            (
                "// chaos-testing-point:v1/subroundEndRound_doEndRoundJobByLeader_delayBroadcastingFinalBlock",
                """chaos.Controller.In_V1_subroundEndRound_doEndRoundJobByLeader_maybeDelayBroadcastingFinalBlock(sr)"""
            )
        ],
        with_import=True
    )

    do_replacements(
        file_path=Path("consensus/spos/bls/v2/subroundBlock.go"),
        replacements=[
            (
                "// chaos-testing-point:v2/subroundBlock_doBlockJob_corruptLeaderSignature",
                """chaos.Controller.In_V2_subroundBlock_doBlockJob_maybeCorruptLeaderSignature(sr, leaderSignature)"""
            ),
            (
                "// chaos-testing-point:v2/subroundBlock_doBlockJob_delayLeaderSignature",
                """chaos.Controller.In_V2_subroundBlock_doBlockJob_maybeDelayLeaderSignature(sr)"""
            ),
            (
                "// chaos-testing-point:v2/subroundBlock_doBlockJob_skipSendingBlock",
                """if chaos.Controller.In_V2_subroundBlock_doBlockJob_shouldSkipSendingBlock(sr) {
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
                """chaos.Controller.In_V1_and_V2_subroundSignature_doSignatureJob_maybeCorruptSignature_whenSingleKey(sr, signatureShare)"""
            ),
            (
                "// chaos-testing-point:v2/subroundSignature_doSignatureJob_corruptSignatureWhenMultiKey",
                """chaos.Controller.In_V1_and_V2_subroundSignature_doSignatureJob_maybeCorruptSignature_whenMultiKey(sr, pk, signatureShare)"""
            )
        ],
        with_import=True
    )

    ensure_no_marker_skipped()


def do_replacements(file_path: Path, replacements: list[Tuple[str, str]], with_import: bool = False):
    print(f"Inserting patches into {file_path}...")

    content = file_path.read_text()

    for marker, patch in replacements:
        content = content.replace(marker, patch)

    if with_import:
        content = add_chaos_import(content)

    file_path.write_text(content)


def add_chaos_import(file_content: str) -> str:
    statement = "import \"github.com/multiversx/mx-chain-go/chaos\""

    if statement in file_content:
        return file_content

    lines = file_content.split("\n")
    lines.insert(1, statement)
    return "\n".join(lines)


def ensure_no_marker_skipped():
    print("Ensuring no marker is skipped...")

    all_source_files = list(Path(".").rglob("*.go"))

    for file in all_source_files:
        content = file.read_text()
        assert "chaos-testing-point" not in content, f"Marker skipped in {file}."

    print("No marker is skipped.")


if __name__ == "__main__":
    main()
