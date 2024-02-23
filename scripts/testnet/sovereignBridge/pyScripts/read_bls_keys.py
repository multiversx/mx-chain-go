import json


def read_and_concatenate_pubkeys(file_path):
    with open(file_path, 'r') as file:
        data = json.load(file)
        initial_nodes = data["initialNodes"]
        pubkeys = [node["pubkey"] for node in initial_nodes]
        concatenated_pubkeys = ' '.join(pubkeys)
        return concatenated_pubkeys


def main():
    file_path = "/home/ubuntu/MultiversX/testnet/node/config/nodesSetup.json"
    print(read_and_concatenate_pubkeys(file_path))


if __name__ == "__main__":
    main()
