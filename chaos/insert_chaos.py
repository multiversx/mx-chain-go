import argparse
import json
from pathlib import Path
from typing import Any, Optional, Tuple


def main():
    parser = argparse.ArgumentParser(description="Utility to insert chaotic behavior into chaos points. To be used for testing, within CI pipelines.")
    parser.add_argument("--config-file", required=True, help="Path of the chaos configuration file")

    args = parser.parse_args()
    config = load_config(args.config_file)
    profile_name = config["selectedProfile"]
    profile = get_profile(config, profile_name)

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
            replacement("shardBlockCreateBlock", out_on_error=["nil", "nil", "ERROR"]),
            replacement("shardBlockProcessBlock", out_on_error=["ERROR"])
        ],
        with_import=True
    )

    do_replacements(
        file_path=Path("process/block/metablock.go"),
        replacements=[
            replacement("metaBlockCreateBlock", out_on_error=["nil", "nil", "ERROR"]),
            replacement("metaBlockProcessBlock", out_on_error=["ERROR"]),
        ],
        with_import=True
    )

    do_replacements(
        file_path=Path("consensus/spos/bls/v1/subroundSignature.go"),
        replacements=[
            replacement("consensusV1SubroundSignatureDoSignatureJobWhenSingleKey", in_consensus_state="sr", in_signature="signatureShare", out_on_error=["false"], out_on_value=["BOOLEAN"]),
            replacement("consensusV1SubroundSignatureDoSignatureJobWhenMultiKey", in_consensus_state="sr", in_node_public_key="pk", in_signature="signatureShare", out_on_error=["false"], out_on_value=["BOOLEAN"]),
        ],
        with_import=True
    )

    do_replacements(
        file_path=Path("consensus/spos/bls/v1/subroundEndRound.go"),
        replacements=[
            replacement("consensusV1SubroundEndRoundCheckSignaturesValidity", in_consensus_state="sr", out_on_error=["spos.ErrInvalidSignature"]),
            replacement("consensusV1SubroundEndRoundDoEndRoundJobByLeaderBeforeBroadcastingFinalBlock", in_consensus_state="sr", out_on_error=["false"], out_on_value=["BOOLEAN"]),
        ],
        with_import=True
    )

    do_replacements(
        file_path=Path("consensus/spos/bls/v2/subroundBlock.go"),
        replacements=[
            replacement("consensusV2SubroundBlockDoBlockJob", in_consensus_state="sr", in_header="header", in_signature="leaderSignature", out_on_error=["false"], out_on_value=["BOOLEAN"]),
        ],
        with_import=True
    )

    do_replacements(
        file_path=Path("consensus/spos/bls/v2/subroundSignature.go"),
        replacements=[
            replacement("consensusV2SubroundSignatureDoSignatureJobWhenSingleKey", in_consensus_state="sr", in_signature="signatureShare", out_on_error=["false"], out_on_value=["BOOLEAN"]),
            replacement("consensusV2SubroundSignatureDoSignatureJobWhenMultiKey", in_consensus_state="sr", in_node_public_key="pk", in_signature="signatureShare", out_on_error=["false"], out_on_value=["BOOLEAN"]),
        ],
        with_import=True
    )

    do_replacements(
        file_path=Path("consensus/spos/bls/v2/subroundEndRound.go"),
        replacements=[
            replacement("consensusV2SubroundEndRoundOnReceivedProof", in_consensus_state="sr"),
            replacement("consensusV2SubroundEndRoundVerifyInvalidSigner", in_consensus_state="sr", out_on_error=["ERROR"]),
            replacement("consensusV2SubroundEndRoundCommitBlock", in_consensus_state="sr", out_on_error=["ERROR"]),
            replacement("consensusV2SubroundEndRoundDoEndRoundJobByNode", in_consensus_state="sr", out_on_error=["false"], out_on_value=["BOOLEAN"]),
            replacement("consensusV2SubroundEndRoundPrepareBroadcastBlockData", in_consensus_state="sr", out_on_error=["ERROR"]),
            replacement("consensusV2SubroundEndRoundWaitForProof", in_consensus_state="sr", out_on_error=["false"], out_on_value=["BOOLEAN"]),
            replacement("consensusV2SubroundEndRoundFinalizeConfirmedBlock", in_consensus_state="sr", out_on_error=["false"], out_on_value=["BOOLEAN"]),
            replacement("consensusV2SubroundEndRoundSendProof", in_consensus_state="sr"),
            replacement("consensusV2SubroundEndRoundShouldSendProof", in_consensus_state="sr", out_on_error=["false"], out_on_value=["BOOLEAN"]),
            replacement("consensusV2SubroundEndRoundAggregateSigsAndHandleInvalidSigners", in_consensus_state="sr", out_on_error=["nil", "nil", "ERROR"]),
            replacement("consensusV2SubroundEndRoundVerifySignature", in_consensus_state="sr", out_on_error=["ERROR"]),
            replacement("consensusV2SubroundEndRoundVerifySignatureBeforeChangeScore", in_consensus_state="sr", in_corruptible=["&decreaseFactor"], out_on_error=["ERROR"]),
            replacement("consensusV2SubroundEndRoundCreateAndBroadcastProof", in_consensus_state="sr", out_on_error=["ERROR"]),
            replacement("consensusV2SubroundEndRoundCheckSignaturesValidity", in_consensus_state="sr", out_on_error=["ERROR"]),
            replacement("consensusV2SubroundEndRoundIsOutOfTime", in_consensus_state="sr", out_on_error=["false"], out_on_value=["BOOLEAN"]),
            replacement("consensusV2SubroundEndRoundRemainingTime", in_consensus_state="sr", out_on_value=["DURATION"]),
            replacement("consensusV2SubroundEndRoundOnReceivedSignature", in_consensus_state="sr", out_on_error=["false"], out_on_value=["BOOLEAN"]),
            replacement("consensusV2SubroundEndRoundCheckReceivedSignatures", in_consensus_state="sr", out_on_error=["false"], out_on_value=["BOOLEAN"]),
            replacement("consensusV2SubroundEndRoundGetNumOfSignaturesCollected", in_consensus_state="sr", out_on_value=["INT"]),
        ],
        with_import=True
    )

    ensure_no_marker_skipped()


def load_config(config_file: Path) -> dict[str, Any]:
    config_file = Path(config_file).expanduser().resolve()

    if not config_file.exists():
        raise FileNotFoundError(f"File not found: {config_file}")

    return json.loads(config_file.read_text())


def get_profile(config: dict[str, Any], profile_name: str) -> dict[str, Any]:
    for profile in config["profiles"]:
        if profile["name"] == profile_name:
            return profile

    raise ValueError(f"Profile not found: {profile_name}")


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
    if not file_path.exists():
        print(f"WARNING: File not found: {file_path}")
        return

    print(f"Inserting patches into {file_path}...")

    content = file_path.read_text()

    for original_text, new_text in replacements:
        if original_text not in content:
            print(f"WARNING: Text not found in {file_path}: {original_text}.")
            continue

        content = content.replace(original_text, new_text)

    if with_import:
        content = add_chaos_import(content)

    file_path.write_text(content)


def replacement(point: str,
                in_consensus_state: str = "",
                in_node_public_key: str = "",
                in_header: str = "",
                in_signature: str = "",
                in_corruptible: Optional[list[str]] = None,
                out_on_error: Optional[list[str]] = None,
                out_on_value: Optional[list[str]] = None) -> Tuple[str, str]:

    chaos_marker = f"// chaos:{point}\n"

    input_initializer_body = f"Name: \"{point}\""

    if in_consensus_state:
        input_initializer_body += f", ConsensusState: {in_consensus_state}"
    if in_node_public_key:
        input_initializer_body += f", NodePublicKey: {in_node_public_key}"
    if in_header:
        input_initializer_body += f", Header: {in_header}"
    if in_signature:
        input_initializer_body += f", Signature: {in_signature}"
    if in_corruptible:
        input_initializer_body += f", Corruptible: []interface{{}}{{{', '.join(in_corruptible)}}}"

    input_text = f"chaos.PointInput{{{input_initializer_body}}}"

    known_output_parameters = {
        "ERROR": "chaosOutput.Error",
        "BOOLEAN": "chaosOutput.Boolean",
        "DURATION": "chaosOutput.Duration",
        "INT": "chaosOutput.NumberInt",
    }

    if out_on_error is None:
        output_on_error_text = "// do nothing"
    elif len(out_on_error) == 0:
        output_on_error_text = "return"
    else:
        out_on_error = [known_output_parameters.get(param, param) for param in out_on_error]
        output_on_error_text = f"return {', '.join(out_on_error)}"

    if out_on_value is None:
        output_on_value_text = "// do nothing"
    elif len(out_on_value) == 0:
        output_on_value_text = "return"
    else:
        out_on_value = [known_output_parameters.get(param, param) for param in out_on_value]
        output_on_value_text = f"return {', '.join(out_on_value)}"

    replacement = f"""chaosOutput := chaos.Controller.HandlePoint({input_text})
    if chaosOutput.Error != nil {{
        {output_on_error_text}
    }}
    if chaosOutput.HasValue {{
        {output_on_value_text}
    }}
    """

    return chaos_marker, replacement


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

        if marker in content:
            print(f"WARNING: Chaos markers still present (skipped) in file: {file}")

    print("No marker is skipped.")


if __name__ == "__main__":
    main()
