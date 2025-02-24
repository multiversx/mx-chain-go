import argparse
import json
from pathlib import Path
from typing import Any, Tuple


def main():
    parser = argparse.ArgumentParser(description="Utility to insert chaotic behavior into chaos points. To be used for testing, within CI pipelines.")
    parser.add_argument("--config-file", required=True, help="Path of the chaos configuration file")

    args = parser.parse_args()
    config = load_config(args.config_file)
    profile_name = config["selectedProfile"]
    profile = config["profiles"][profile_name]

    alter_code_constants(profile)

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
        file_path=Path("process/block/metablock.go"),
        replacements=[
            (
                "// chaos:metaBlockCreateBlock",
                """if chaos.Controller.HandlePoint(chaos.PointInput{Name: \"metaBlockCreateBlock\"}) != nil {
        return nil, nil, chaos.ErrChaoticBehavior
	}"""
            ),
            (
                "// chaos:metaBlockProcessBlock",
                """if chaos.Controller.HandlePoint(chaos.PointInput{Name: \"metaBlockProcessBlock\"}) != nil {
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


def load_config(config_file: Path) -> dict[str, Any]:
    config_file = Path(config_file).expanduser().resolve()

    if not config_file.exists():
        raise FileNotFoundError(f"File not found: {config_file}")

    return json.loads(config_file.read_text())


def alter_code_constants(profile: dict[str, Any]):
    code_constants: dict[str, Any] = profile["codeConstants"]

    constants_replacements_by_file: dict[str, list[tuple[str, str]]] = {
        "common/constants.go": [],
    }

    for key, value in code_constants.items():
        if key == "InvalidMessageBlacklistDuration":
            constants_replacements_by_file["common/constants.go"].append((
                "const InvalidMessageBlacklistDuration = time.Second * 3600",
                f"const InvalidMessageBlacklistDuration = time.Second * {value}"
            ))
        elif key == "PublicKeyBlacklistDuration":
            constants_replacements_by_file["common/constants.go"].append((
                "const PublicKeyBlacklistDuration = time.Second * 7200",
                f"const PublicKeyBlacklistDuration = time.Second * {value}"
            ))
        elif key == "InvalidSigningBlacklistDuration":
            constants_replacements_by_file["common/constants.go"].append((
                "const InvalidSigningBlacklistDuration = time.Second * 7200",
                f"const InvalidSigningBlacklistDuration = time.Second * {value}"
            ))

    for file, replacements in constants_replacements_by_file.items():
        do_replacements(Path(file), replacements)


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
