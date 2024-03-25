import json
import os
import shutil
import sys


def copy_contract_wasm(file_path: str, wasm_path: str):
    try:
        real_wasm_path = os.path.expanduser(wasm_path)
        if os.path.exists(real_wasm_path):
            os.remove(real_wasm_path)

        shutil.copy2(file_path, real_wasm_path)
        print(f"{file_path} copied successfully.")
    except FileNotFoundError:
        print(f"File {file_path} not found.")
    except PermissionError:
        print("Permission denied.")
    except Exception as e:
        print(f"An error occurred: {e}")


def main():
    # input arguments
    esdt_safe_path = sys.argv[1]
    fee_market_path = sys.argv[2]
    multisig_path = sys.argv[3]

    current_path = os.getcwd()
    project = 'mx-chain-go'
    index = current_path.find(project)
    project_path = current_path[:index]
    contracts_path = os.path.join(project_path, 'mx-sovereign-sc')

    copy_contract_wasm(contracts_path + "/esdt-safe/output/esdt-safe.wasm", esdt_safe_path)
    copy_contract_wasm(contracts_path + "/fee-market/output/fee-market.wasm", fee_market_path)
    copy_contract_wasm(contracts_path + "/multisigverifier/output/multisigverifier.wasm", multisig_path)


if __name__ == "__main__":
    main()
