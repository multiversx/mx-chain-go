
import sys
from pathlib import Path
from typing import List

from multiversx_sdk import (Address, ProxyNetworkProvider,  # type: ignore
                            Transaction, TransactionComputer,
                            TransactionsFactoryConfig, UserPEM, UserSigner)

CHAIN_ID = "local-testnet"


def main(cli_args: List[str]):
    factory_config = TransactionsFactoryConfig(chain_id=CHAIN_ID)
    factory_config.issue_cost = 5000000000000000000

    network_provider = ProxyNetworkProvider("http://localhost:9500")
    transaction_computer = TransactionComputer()

    alice_pem = UserPEM.from_file(Path("~/MultiversX/testnet/filegen/output/walletKey.pem").expanduser())
    alice_secret_key = alice_pem.secret_key
    alice_address = Address.from_bech32(alice_pem.public_key.to_address("erd").to_bech32())
    alice_signer = UserSigner(alice_secret_key)

    def get_nonce():
        return network_provider.get_account(alice_address).nonce

    def sign_transaction(transaction: Transaction, signer: UserSigner):
        bytes_for_signing = transaction_computer.compute_bytes_for_signing(transaction)
        transaction.signature = signer.sign(bytes_for_signing)

    def broadcast_transaction(transaction: Transaction):
        hash = network_provider.send_transaction(transaction)
        print(f"Broadcasted transaction with nonce {transaction.nonce}. Hash: {hash}")

    def broadcast_command(data: bytes):
        print(f"Broadcasting command: {data}")

        transaction = Transaction(
            sender=alice_address.bech32(),
            receiver=alice_address.bech32(),
            gas_limit=1000000,
            chain_id=CHAIN_ID,
            value=0,
            gas_price=2000000000,
            data=data,
            version=2,
        )

        transaction.nonce = get_nonce()
        sign_transaction(transaction, alice_signer)
        broadcast_transaction(transaction)

    num_accounts = 24
    txs_per_account = 50000
    max_transactions_per_block = 80000
    max_batch_size_per_participant = int(max_transactions_per_block / num_accounts)

    data = "ext_run_scenario_move_balances@{:08x}@{:08x}@{:08x}@{:08x}".format(num_accounts, txs_per_account, max_transactions_per_block, max_batch_size_per_participant).encode()
    broadcast_command(data)


def my_hex(value: int):
    return '{:08x}'.format(value)


if __name__ == "__main__":
    main(sys.argv[1:])
