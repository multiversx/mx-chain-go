
import re
import sys
from argparse import ArgumentParser
from pathlib import Path
from typing import List

from multiversx_sdk import (Address, ProxyNetworkProvider,  # type: ignore
                            Transaction, TransactionComputer,
                            TransactionsFactoryConfig,
                            TransferTransactionsFactory, UserSigner)


class Sender:
    def __init__(self, signer: UserSigner):
        self.signer = signer
        self.address = signer.get_pubkey().to_address("erd")
        self.nonce = 0

    def sync_nonce(self, network_provider: ProxyNetworkProvider):
        self.nonce = network_provider.get_account(self.address).nonce


def main(cli_args: List[str]):
    parser = ArgumentParser()
    parser.add_argument("--proxy", required=True)
    parser.add_argument("--wallets", required=True)
    parser.add_argument("--indices", required=True)
    parser.add_argument("--count", default=42)
    args = parser.parse_args(cli_args)

    proxy_url = args.proxy
    wallets_pem_path = Path(args.wallets)
    indices = parse_sender_indices(args.indices)

    network_provider = ProxyNetworkProvider(proxy_url)
    transaction_computer = TransactionComputer()
    chain_id = network_provider.get_network_config().chain_id
    factory_config = TransactionsFactoryConfig(chain_id=chain_id)
    transfers_factory = TransferTransactionsFactory(factory_config)

    senders: List[Sender] = []

    for index in indices:
        signer = UserSigner.from_pem_file(wallets_pem_path, index)
        senders.append(Sender(signer))

    for sender in senders:
        sender.sync_nonce(network_provider)

    transactions: List[Transaction] = []

    for sender in senders:
        for _ in range(int(args.count)):
            transaction = transfers_factory.create_transaction_for_native_token_transfer(
                sender=sender.address,
                receiver=sender.address,
                native_amount=1,
            )

            transaction.nonce = sender.nonce
            sender.nonce += 1

            bytes_for_signing = transaction_computer.compute_bytes_for_signing(transaction)
            transaction.signature = sender.signer.sign(bytes_for_signing)

            transactions.append(transaction)

    print(f"Sending {len(transactions)} transactions")
    num_txs, _ = network_provider.send_transactions(transactions)
    print(f"Sent: {num_txs}")


def parse_sender_indices(indices: str) -> List[int]:
    if not indices:
        return []

    # Handle specific indices
    try:
        parts = indices.split(",")
        return [int(part) for part in parts]
    except Exception:
        pass

    # Handle ranges. E.g. 0:4.
    try:
        parts = indices.split(":")
        return list(range(int(parts[0]), int(parts[1]) + 1))
    except Exception:
        pass

    raise Exception(f"Cannot parse indices: {indices}")


if __name__ == "__main__":
    main(sys.argv[1:])
