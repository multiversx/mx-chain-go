import argparse
import json
from pathlib import Path
from typing import Any


def main():
    parser = argparse.ArgumentParser(description="Configuration utility for chaos testing, to be used within CI pipelines.")

    subparsers = parser.add_subparsers()

    subparser = subparsers.add_parser("select-profile", help="Select a profile")
    subparser.add_argument("--config-file", required=True, help="Path of the chaos configuration file to be altered")
    subparser.add_argument("--profile", required=True, help="The profile to select")
    subparser.set_defaults(func=select_profile)

    subparser = subparsers.add_parser("enable", help="Enable a failure")
    subparser.add_argument("--config-file", required=True, help="Path of the chaos configuration file to be altered")
    subparser.add_argument("--profile", required=True, help="The profile to alter")
    subparser.add_argument("--failure", required=True, help="The failure to enable")
    subparser.set_defaults(func=enable_failure)

    subparser = subparsers.add_parser("disable", help="Disable a failure")
    subparser.add_argument("--config-file", required=True, help="Path of the chaos configuration file to be altered")
    subparser.add_argument("--profile", required=True, help="The profile to alter")
    subparser.add_argument("--failure", required=True, help="The failure to disable")
    subparser.set_defaults(func=disable_failure)

    # Parse arguments
    args = parser.parse_args()

    if hasattr(args, "func"):
        args.func(args)
    else:
        parser.print_help()


def select_profile(args: argparse.Namespace):
    config = load_config(args.config_file)
    config["selectedProfile"] = args.profile
    store_config(config, args.config_file)


def enable_failure(args: argparse.Namespace):
    toggle_failure(args, True)


def disable_failure(args: argparse.Namespace):
    toggle_failure(args, False)


def toggle_failure(args: argparse.Namespace, enabled: bool):
    config = load_config(args.config_file)

    failure = get_failure(config, args.profile, args.failure)
    failure["enabled"] = enabled

    store_config(config, args.config_file)


def get_failure(config: dict[str, Any], profile_name: str, failure_name: str) -> dict[str, Any]:
    profile = get_profile(config, profile_name)

    for failure in profile["failures"]:
        if failure["name"] == failure_name:
            return failure

    raise ValueError(f"Failure not found: {failure_name}")


def get_profile(config: dict[str, Any], profile_name: str) -> dict[str, Any]:
    for profile in config["profiles"]:
        if profile["name"] == profile_name:
            return profile

    raise ValueError(f"Profile not found: {profile_name}")


def load_config(config_file: Path) -> dict[str, Any]:
    config_file = Path(config_file).expanduser().resolve()

    if not config_file.exists():
        raise FileNotFoundError(f"File not found: {config_file}")

    return json.loads(config_file.read_text())


def store_config(config: dict[str, Any], config_file: Path):
    config_file = Path(config_file).expanduser().resolve()
    config_file.write_text(json.dumps(config, indent=4))


if __name__ == "__main__":
    main()
