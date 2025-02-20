from pathlib import Path
from typing import Tuple


def main():
    # Insert setup points:

    do_replacements(
        file_path=Path("cmd/node/main.go"),
        replacements=[
            (
                "// chaos:setup",
                """errChaosSetup := chaos.Controller.Setup()
    if errChaosSetup != nil {
        return errChaosSetup
    }
"""
            ),
            (
                "// chaos:node_main_handleNodeConfig",
                """chaos.Controller.HandleNodeConfig(cfgs)"""
            )
        ],
        with_import=True
    )

    do_replacements(
        file_path=Path("node/nodeRunner.go"),
        replacements=[
            (
                "// chaos:nodeRunner_handleNode",
                """chaos.Controller.HandleNode(currentNode)"""
            )
        ],
        with_import=True
    )

    # Insert chaos points:

    do_replacements(
        file_path=Path("process/block/shardblock.go"),
        replacements=[
            (
                "// chaos:shardBlockCreateBlock",
                """if chaos.Controller.HandlePoint(chaos.PointInput{Name: \"shardBlockCreateBlock\"}) != nil {
        return nil, nil, chaos.ErrChaoticBehavior
	}"""
            ),
            (
                "// chaos:shardBlockProcessBlock",
                """if chaos.Controller.HandlePoint(chaos.PointInput{Name: \"shardBlockProcessBlock\"}) != nil {
        return chaos.ErrChaoticBehavior
	}"""
            ),
        ],
        with_import=True
    )

    do_replacements(
        file_path=Path("consensus/spos/bls/v1/subroundSignature.go"),
        replacements=[
            (
                "// chaos:consensusV1SubroundSignatureDoSignatureJobWhenSingleKey",
                """chaos.Controller.HandlePoint(chaos.PointInput{Name: \"consensusV1SubroundSignatureDoSignatureJobWhenSingleKey\", ConsensusState: sr, Signature: signatureShare})"""
            ),
            (
                "// chaos:consensusV1SubroundSignatureDoSignatureJobWhenMultiKey",
                """chaos.Controller.HandlePoint(chaos.PointInput{Name: \"consensusV1SubroundSignatureDoSignatureJobWhenMultiKey\", ConsensusState: sr, NodePublicKey: pk, Signature: signatureShare})"""
            )
        ],
        with_import=True
    )

    do_replacements(
        file_path=Path("consensus/spos/bls/v1/subroundEndRound.go"),
        replacements=[
            (
                "// chaos:consensusV1SubroundEndRoundCheckSignaturesValidity",
                """if chaos.Controller.HandlePoint(chaos.PointInput{Name: \"consensusV1SubroundEndRoundCheckSignaturesValidity\", ConsensusState: sr}) != nil {
		return spos.ErrInvalidSignature
	}"""
            ),
            (
                "// chaos:consensusV1SubroundEndRoundDoEndRoundJobByLeaderBeforeBroadcastingFinalBlock",
                """chaos.Controller.HandlePoint(chaos.PointInput{Name: \"consensusV1SubroundEndRoundDoEndRoundJobByLeaderBeforeBroadcastingFinalBlock\", ConsensusState: sr})"""
            )
        ],
        with_import=True
    )

    do_replacements(
        file_path=Path("consensus/spos/bls/v2/subroundBlock.go"),
        replacements=[
            (
                "// chaos:consensusV2SubroundBlockDoBlockJob",
                """errChaos := chaos.Controller.HandlePoint(chaos.PointInput{Name: \"consensusV2SubroundBlockDoBlockJob\", ConsensusState: sr, Header: header, Signature: leaderSignature})
    if errChaos == chaos.ErrEarlyReturn {
        return false
    }"""
            )
        ],
        with_import=True
    )

    do_replacements(
        file_path=Path("consensus/spos/bls/v2/subroundSignature.go"),
        replacements=[
            (
                "// chaos:consensusV2SubroundSignatureDoSignatureJobWhenSingleKey",
                """chaos.Controller.HandlePoint(chaos.PointInput{Name: \"consensusV2SubroundSignatureDoSignatureJobWhenSingleKey\", ConsensusState: sr, Signature: signatureShare})"""
            ),
            (
                "// chaos:consensusV2SubroundSignatureDoSignatureJobWhenMultiKey",
                """chaos.Controller.HandlePoint(chaos.PointInput{Name: \"consensusV2SubroundSignatureDoSignatureJobWhenMultiKey\", ConsensusState: sr, NodePublicKey: pk, Signature: signatureShare})"""
            )
        ],
        with_import=True
    )

    ensure_no_marker_skipped()


def do_replacements(file_path: Path, replacements: list[Tuple[str, str]], with_import: bool = False):
    print(f"Inserting patches into {file_path}...")

    content = file_path.read_text()

    for marker, patch in replacements:
        assert marker in content, f"Marker not found in {file_path}: {marker}."
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

    marker = "// chaos:"
    all_source_files = list(Path(".").rglob("*.go"))

    for file in all_source_files:
        content = file.read_text()
        assert marker not in content, f"Marker is skipped in {file}."

    print("No marker is skipped.")


if __name__ == "__main__":
    main()
