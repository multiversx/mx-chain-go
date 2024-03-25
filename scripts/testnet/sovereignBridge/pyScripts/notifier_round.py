import os
import re
import sys

import requests


def update_notarization_round(lines, section, identifier, round) -> []:
    updated_lines = []
    section_found = False

    for line in lines:
        if line.startswith("[" + section + "]"):
            section_found = True
        if section_found and identifier in line:
            line = re.sub(r"(MainChainNotarizationStartRound\s*=\s*)\d+", r"\g<1>" + str(round), line)
            section_found = False
        updated_lines.append(line)

    return updated_lines


def get_current_round(proxy, shard):
    params = {
        "size": 1,
        "shard": shard,
        "fields": "round"
    }

    api = proxy.replace("gateway", "api")
    response = requests.get(api + f"/blocks?size=1&shard={shard}&fields=round")

    if response.status_code == 200:
        data = response.json()
        if data:
            return data[0]["round"]
        else:
            print("No data returned from the API.")
            return -1
    else:
        print("Failed to retrieve data from the API. Status code:", response.status_code)
        return -1


def main():
    # input arguments
    proxy = sys.argv[1]
    shard = sys.argv[2]

    current_path = os.getcwd()
    project = 'mx-chain-go'
    index = current_path.find(project)
    project_path = current_path[:index + len(project)]
    toml_path = project_path + "/cmd/sovereignnode/config/sovereignConfig.toml"

    current_round = get_current_round(proxy, shard)
    if current_round == -1:
        return

    with open(toml_path, 'r') as file:
        lines = file.readlines()

    updated_lines = update_notarization_round(lines, "MainChainNotarization", "MainChainNotarizationStartRound",
                                              current_round)

    with open(toml_path, 'w') as file:
        file.writelines(updated_lines)


if __name__ == "__main__":
    main()
