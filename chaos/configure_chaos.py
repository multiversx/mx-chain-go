import argparse
import json
from pathlib import Path
from typing import Any


def main():
    parser = argparse.ArgumentParser(description="Configuration utility for chaos testing, to be used within CI pipelines.")

    subparsers = parser.add_subparsers()

    subparser = subparsers.add_parser("enable", help="Enable a failure")
    subparser.add_argument("--config-file", required=True, help="Path of the chaos configuration file to be altered")
    subparser.add_argument("--failure", required=True, help="The failure to enable")
    subparser.set_defaults(func=enable_failure)

    subparser = subparsers.add_parser("disable", help="Disable a failure")
    subparser.add_argument("--config-file", required=True, help="Path of the chaos configuration file to be altered")
    subparser.add_argument("--failure", required=True, help="The failure to disable")
    subparser.set_defaults(func=disable_failure)

    # Parse arguments
    args = parser.parse_args()

    if hasattr(args, "func"):
        args.func(args)
    else:
        parser.print_help()


def enable_failure(args: argparse.Namespace):
    toggle_failure(args, True)


def disable_failure(args: argparse.Namespace):
    toggle_failure(args, False)


def toggle_failure(args: argparse.Namespace, enabled: bool):
    config_file = Path(args.config_file).expanduser().resolve()

    if not config_file.exists():
        raise FileNotFoundError(f"File not found: {config_file}")

    config = load_config(config_file)
    do_toggle_failure(config, args.failure, enabled)
    store_config(config, config_file)


def do_toggle_failure(config: dict[str, Any], failure: str, enabled: bool):
    for item in config["failures"]:
        if item["name"] == failure:
            item["enabled"] = enabled
            return

    raise ValueError(f"Failure not found: {failure}")


def load_config(config_file: Path) -> dict[str, Any]:
    return json.loads(config_file.read_text())


def store_config(config: dict[str, Any], config_file: Path):
    config_file.write_text(json.dumps(config, indent=4))


if __name__ == "__main__":
    main()
