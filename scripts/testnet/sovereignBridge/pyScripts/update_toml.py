import os
import re
import sys


def update_subscribed_addresses(lines, section, identifier, main_chain_address) -> []:
    updated_lines = []
    section_found = False

    for line in lines:
        if line.startswith("[" + section + "]"):
            section_found = True
        if section_found and identifier in line:
            line = re.sub(r'(Addresses\s*=\s*\[)[^\]]*', r'\1' + f"\"{main_chain_address}\"", line)
            section_found = False
        updated_lines.append(line)

    return updated_lines


def update_key(lines, key, value) -> []:
    updated_lines = []

    for line in lines:
        if key in line:
            line = re.sub(rf'{re.escape(key)}\s*=\s*".*?"', f'{key} = "{value}"', line)
        updated_lines.append(line)

    return updated_lines


def enable_key(lines, section):
    updated_lines = []
    section_found = False

    for line in lines:
        if line.startswith("[" + section + "]"):
            section_found = True
        if section_found and "Enabled" in line:
            line = re.sub(r'(Enabled\s*=\s*)\w+', r'\1true', line)
            section_found = False
        updated_lines.append(line)

    return updated_lines


def update_sovereign_config(file_path, main_chain_address, sovereign_chain_address):
    with open(file_path, 'r') as file:
        lines = file.readlines()

    updated_lines = update_subscribed_addresses(lines, "OutgoingSubscribedEvents", "deposit", sovereign_chain_address)
    updated_lines = update_subscribed_addresses(updated_lines, "NotifierConfig", "deposit", main_chain_address)
    updated_lines = update_subscribed_addresses(updated_lines, "NotifierConfig", "execute", main_chain_address)
    updated_lines = enable_key(updated_lines, "OutGoingBridge")
    updated_lines = enable_key(updated_lines, "NotifierConfig")

    with open(file_path, 'w') as file:
        file.writelines(updated_lines)


def update_esdt_prefix(file_path, esdt_prefix):
    with open(file_path, 'r') as file:
        lines = file.readlines()

    updated_lines = update_key(lines, "ESDTPrefix", esdt_prefix)

    with open(file_path, 'w') as file:
        file.writelines(updated_lines)


def update_transfer_and_execute_address(file_path, address):
    with open(file_path, 'r') as file:
        lines = file.readlines()

    for i, line in enumerate(lines):
        if "TransferAndExecuteByUserAddresses" in line:
            lines[i + 1] = re.sub(r'".*?"', f'"{address}"', lines[i + 1])

    with open(file_path, 'w') as file:
        file.writelines(lines)


def update_node_configs(config_path, esdt_prefix, sovereign_chain_address):
    update_esdt_prefix(config_path + "/systemSmartContractsConfig.toml", esdt_prefix)
    update_transfer_and_execute_address(config_path + "/config.toml", sovereign_chain_address)


def update_main_chain_elastic_url(lines, section, key, value):
    updated_lines = []
    section_found = False

    for line in lines:
        if line.startswith("[" + section + "]"):
            section_found = True
        if section_found and f"{key} =" in line:
            line = re.sub(rf'({re.escape(key)}\s*=\s*)".*?"', rf'\1"{value}"', line)
            section_found = False
        updated_lines.append(line)

    return updated_lines


def update_external_config(file_path, main_chain_elastic):
    with open(file_path, 'r') as file:
        lines = file.readlines()

    updated_lines = update_main_chain_elastic_url(lines, "MainChainElasticSearchConnector", "URL", main_chain_elastic)

    with open(file_path, 'w') as file:
        file.writelines(updated_lines)


def main():
    # input arguments
    main_chain_address = sys.argv[1]
    sovereign_chain_address = sys.argv[2]
    esdt_prefix = sys.argv[3]
    main_chain_elastic = sys.argv[4]

    current_path = os.getcwd()
    project = 'mx-chain-go'
    index = current_path.find(project)
    project_path = current_path[:index + len(project)]
    toml_path = project_path + "/cmd/sovereignnode/config/sovereignConfig.toml"
    update_sovereign_config(toml_path, main_chain_address, sovereign_chain_address)

    config_path = project_path + "/cmd/node/config"
    update_node_configs(config_path, esdt_prefix, sovereign_chain_address)

    external_path = project_path + "/cmd/node/config/external.toml"
    update_external_config(external_path, main_chain_elastic)


if __name__ == "__main__":
    main()
