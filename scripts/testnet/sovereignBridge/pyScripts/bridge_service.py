import os
import re
import subprocess
import sys
from datetime import datetime


def update_env(lines, identifier, value) -> []:
    updated_lines = []

    for line in lines:
        if line.startswith(identifier):
            line = re.sub(r'"(.*?)"', f'"{value}"', line)

        updated_lines.append(line)

    return updated_lines


def build_and_run_server(server_path):
    os.chdir(server_path)

    build_command = "go build"
    run_service = "screen -L -Logfile sovereignBridgeService.txt -d -m -S sovereignBridgeService ./server"

    build_process = subprocess.run(build_command, shell=True, capture_output=True, text=True)
    if build_process.returncode == 0:
        print("Go build successful.")
    else:
        print("Error during Go build.")
        return

    # TODO start terminal with server app
    run_process = subprocess.run(run_service, shell=True, capture_output=True, text=True)
    if run_process.returncode == 0:
        print("Bridge service running...")
    else:
        print("Error running bridge service.")
        return


def main():
    # input arguments
    wallet = sys.argv[1]
    proxy = sys.argv[2]
    esdt_safe_address = sys.argv[3]
    multisig_address = sys.argv[4]

    current_path = os.getcwd()
    project = 'mx-chain-go'
    index = current_path.find(project)
    project_path = current_path[:index]
    bridge_service_path = os.path.join(project_path, 'mx-chain-sovereign-bridge-go')
    server_path = bridge_service_path + "/server/cmd/server"
    env_path = server_path + "/.env"

    with open(env_path, 'r') as file:
        lines = file.readlines()

    updated_lines = update_env(lines, "WALLET_PATH", os.path.expanduser(wallet))
    updated_lines = update_env(updated_lines, "MULTIVERSX_PROXY", os.path.expanduser(proxy))
    updated_lines = update_env(updated_lines, "MULTISIG_SC_ADDRESS", multisig_address)
    updated_lines = update_env(updated_lines, "ESDT_SAFE_SC_ADDRESS", esdt_safe_address)
    updated_lines = update_env(updated_lines, "CERT_FILE", os.path.expanduser("~/MultiversX/testnet/node/config/certificate.crt"))
    updated_lines = update_env(updated_lines, "CERT_PK_FILE", os.path.expanduser("~/MultiversX/testnet/node/config/private_key.pem"))

    with open(env_path, 'w') as file:
        file.writelines(updated_lines)

    build_and_run_server(server_path)


if __name__ == "__main__":
    main()
