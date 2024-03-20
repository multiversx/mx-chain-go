import json
import os


def read_and_concatenate_pubkeys(file_path):
    with open(file_path, 'r') as file:
        data = json.load(file)
        initial_nodes = data["initialNodes"]
        pubkeys = [f"0x{node['pubkey']}" for node in initial_nodes]
        concatenated_pubkeys = ' '.join(pubkeys)
        return concatenated_pubkeys


def main():
    file_path = "~/MultiversX/testnet/node/config/nodesSetup.json"
    print(read_and_concatenate_pubkeys(os.path.expanduser(file_path)))


if __name__ == "__main__":
    main()
