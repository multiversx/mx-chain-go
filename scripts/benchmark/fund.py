
import re
import sys
from argparse import ArgumentParser
from pathlib import Path
from typing import List

from multiversx_sdk import (Address, ProxyNetworkProvider,  # type: ignore
                            Transaction, TransactionComputer,
                            TransactionsFactoryConfig,
                            TransferTransactionsFactory, UserSigner)


def main(cli_args: List[str]):
    parser = ArgumentParser()
    parser.add_argument("--proxy", required=True)
    parser.add_argument("--sponsor", required=True)
    parser.add_argument("--receivers", required=True)
    parser.add_argument("--value", default="1000000000000000000")
    args = parser.parse_args(cli_args)

    proxy_url = args.proxy
    sponsor_pem_path = Path(args.sponsor)
    receivers_path = Path(args.receivers)
    value = int(args.value)

    network_provider = ProxyNetworkProvider(proxy_url)
    transaction_computer = TransactionComputer()
    chain_id = network_provider.get_network_config().chain_id
    factory_config = TransactionsFactoryConfig(chain_id=chain_id)
    transfers_factory = TransferTransactionsFactory(factory_config)

    sponsor_signer = UserSigner.from_pem_file(sponsor_pem_path.expanduser())
    sponsor_address = sponsor_signer.get_pubkey().to_address("erd")
    sponsor_nonce = network_provider.get_account(sponsor_address).nonce

    receivers = read_receivers(receivers_path)
    print(f"Found {len(receivers)} receivers")

    transactions: List[Transaction] = []

    for receiver in receivers:
        transaction = transfers_factory.create_transaction_for_native_token_transfer(
            sender=sponsor_address,
            receiver=receiver,
            native_amount=value,
        )

        transaction.nonce = sponsor_nonce
        sponsor_nonce += 1

        bytes_for_signing = transaction_computer.compute_bytes_for_signing(transaction)
        transaction.signature = sponsor_signer.sign(bytes_for_signing)

        transactions.append(transaction)

    print(f"Sending {len(transactions)} transactions")
    num_txs, _ = network_provider.send_transactions(transactions)
    print(f"Sent: {num_txs}")


def read_receivers(receivers_path: Path) -> List[Address]:
    lines = receivers_path.read_text().splitlines()
    pattern = re.compile(r"erd1[a-z0-9]{58}")
    addresses = [match.group() for line in lines for match in pattern.finditer(line)]
    addresses = list(dict.fromkeys(addresses))
    return [Address.from_bech32(address) for address in addresses]


if __name__ == "__main__":
    main(sys.argv[1:])
