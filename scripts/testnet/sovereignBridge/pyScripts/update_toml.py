import os
import sys
import re


def update_toml(lines, section, identifier, main_chain_address) -> []:
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


def main():
    # input arguments
    main_chain_address = sys.argv[1]
    sovereign_chain_address = sys.argv[2]

    current_path = os.getcwd()
    project = 'mx-chain-go'
    index = current_path.find(project)
    project_path = current_path[:index + len(project)]
    toml_path = project_path + "/cmd/sovereignnode/config/sovereignConfig.toml"

    with open(toml_path, 'r') as file:
        lines = file.readlines()

    updated_lines = update_toml(lines, "OutgoingSubscribedEvents", "deposit", sovereign_chain_address)
    updated_lines = update_toml(updated_lines, "NotifierConfig", "deposit", main_chain_address)
    updated_lines = update_toml(updated_lines, "NotifierConfig", "execute", main_chain_address)

    with open(toml_path, 'w') as file:
        file.writelines(updated_lines)


if __name__ == "__main__":
    main()
