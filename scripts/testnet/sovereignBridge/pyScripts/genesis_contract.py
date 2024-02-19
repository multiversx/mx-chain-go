import json
import os
import shutil
import sys


def copy_wasm_in_project(file_path: str, esdt_safe_path: str):
    try:
        if os.path.exists(file_path):
            os.remove(file_path)

        shutil.copy2(esdt_safe_path, file_path)
        print(f"esdt-safe wasm file copied successfully.")
    except FileNotFoundError:
        print("File not found.")
    except PermissionError:
        print("Permission denied.")
    except Exception as e:
        print(f"An error occurred: {e}")


def push_genesis_contract(file_path: str, owner_address: str):
    try:
        with open(file_path, 'r') as file:
            data = json.load(file)

        genesis_contract = {
            "owner": owner_address,
            "filename": "./config/genesisContracts/esdt-safe.wasm",
            "init-parameters": "",
            "vm-type": "0500",
            "type": "esdt",
            "version": "0.0.*"
        }

        found = False
        for item in data:
            if item["filename"] == genesis_contract["filename"]:
                item.update(genesis_contract)
                found = True
                break
        if not found:
            data.append(genesis_contract)

        with open(file_path, 'w') as file:
            json.dump(data, file, indent=2)

        with open(file_path, 'a') as file:
            file.write('\n')

        print(f"esdt-safe genesis contract pushed successfully")
    except Exception as e:
        print(f"An error occurred: {e}")


def main():
    # input arguments
    esdt_safe_path = sys.argv[1]
    owner_address = sys.argv[2]

    current_path = os.getcwd()
    project = 'mx-chain-go'
    index = current_path.find(project)
    project_path = current_path[:index + len(project)]

    wasm_path = project_path + "/cmd/node/config/genesisContracts/esdt-safe.wasm"
    copy_wasm_in_project(wasm_path, esdt_safe_path)

    json_path = project_path + "/cmd/node/config/genesisSmartContracts.json"
    push_genesis_contract(json_path, owner_address)


if __name__ == "__main__":
    main()
