import base64
import heapq
import json
from typing import Any

import requests

MEMPOOL_URL = "https://api.multiversx.com/pool"
BULK_GET_SIZE = 10_000
BULK_GET_INDEX_ERROR_TOLERANCE = int(BULK_GET_SIZE / 10)

MIN_GAS_LIMIT = 50_000
GAS_LIMIT_PER_BYTE = 1_500
GAS_PRICE_MODIFIER = 0.01
SHARDS = [0, 1, 2]

# A few blocks worth of gas.
GAS_BUCKET_SIZE = 6_000_000_000
FAST_BUCKET_INDEX = 2
FASTER_BUCKET_INDEX = 0


class MempoolTransaction:
    def __init__(self, hash: str, sender: str, sender_shard: int, nonce: int, data: str, gas_price: int, gas_limit: int, function: str):
        self.hash = hash
        self.sender = sender
        self.sender_shard = sender_shard
        self.nonce = nonce
        self.data = data
        self.gas_price = gas_price
        self.gas_limit = gas_limit
        self.function = function

        data_cost = MIN_GAS_LIMIT + len(data) * GAS_LIMIT_PER_BYTE
        execution_cost = gas_limit - data_cost

        self.initially_paid_fee = data_cost * gas_price + execution_cost * gas_price * GAS_PRICE_MODIFIER
        self.ppu = int(self.initially_paid_fee // gas_limit)

    @classmethod
    def from_record(cls, record: dict[str, Any]):
        hash = record["txHash"]
        sender = record["sender"]
        sender_shard = record["senderShard"]
        nonce = record["nonce"]
        data = base64.b64decode(record["data"]).decode("utf-8") if "data" in record else ""
        gas_price = record["gasPrice"]
        gas_limit = record["gasLimit"]
        function = record.get("function", "")

        return cls(hash, sender, sender_shard, nonce, data, gas_price, gas_limit, function)

    def takes_precedence_over(self, other_transaction: "MempoolTransaction") -> bool:
        if self.sender == other_transaction.sender:
            return self.nonce < other_transaction.nonce

        if self.ppu != other_transaction.ppu:
            return self.ppu > other_transaction.ppu

        if self.gas_limit != other_transaction.gas_limit:
            return self.gas_limit > other_transaction.gas_limit

        return self.hash < other_transaction.hash

    def __lt__(self, other: "MempoolTransaction") -> bool:
        # "Less is more" (we're doing a max heap).
        return self.takes_precedence_over(other)


class GasBucket:
    def __init__(self, index: int):
        self.index = index
        self.gas_begin = index * GAS_BUCKET_SIZE
        self.gas_end = self.gas_begin + GAS_BUCKET_SIZE
        self.gas_accumulated = 0
        self.ppu_begin = 0
        self.ppu_end = 0
        self.num_transactions = 0

    def add_transaction(self, transaction: MempoolTransaction) -> bool:
        self.gas_accumulated += transaction.gas_limit
        self.num_transactions += 1

        if not self.ppu_begin:
            # This is the first transaction added to the bucket (best transaction).
            self.ppu_begin = transaction.ppu

        if self.is_full():
            self.ppu_end = transaction.ppu

            # We don't accept other transactions in the bucket.
            return False

        # We accept other transactions in the bucket.
        return True

    def is_full(self) -> bool:
        return self.gas_accumulated >= GAS_BUCKET_SIZE


def main():
    transactions = get_all_mempool_transactions()
    metadata_by_shard: list[dict[str, int]] = []

    for shard in SHARDS:
        print("=" * 80)

        shard_transactions = [transaction for transaction in transactions if transaction.sender_shard == shard]

        print(f"Shard {shard}, {len(shard_transactions)} transactions")
        sorted_transactions = sort_transactions(shard_transactions)
        buckets = distribute_transactions_into_gas_buckets(sorted_transactions)

        fast_bucket = buckets[FAST_BUCKET_INDEX] if len(buckets) > FAST_BUCKET_INDEX else None
        faster_bucket = buckets[FASTER_BUCKET_INDEX] if len(buckets) > FASTER_BUCKET_INDEX else None

        metadata = {
            "fast": fast_bucket.ppu_end if fast_bucket else 0,
            "faster": faster_bucket.ppu_end if faster_bucket else 0
        }

        metadata_by_shard.append(metadata)

    metadata_by_shard_json = json.dumps(metadata_by_shard, indent=4)
    print("=" * 80)
    print("=" * 80)
    print("=" * 80)
    print(metadata_by_shard_json)


def get_all_mempool_transactions() -> list[MempoolTransaction]:
    transactions_by_hash: dict[str, MempoolTransaction] = {}

    from_index = 0

    while True:
        transactions_page = get_mempool_transactions_page(from_index, BULK_GET_SIZE)

        for transaction in transactions_page:
            if transaction.hash not in transactions_by_hash:
                transactions_by_hash[transaction.hash] = transaction

        if len(transactions_page) < BULK_GET_SIZE:
            break

        from_index = from_index + len(transactions_page) - BULK_GET_INDEX_ERROR_TOLERANCE

    transactions = list(transactions_by_hash.values())
    print(f"Got {len(transactions)} transactions")

    return transactions


def get_mempool_transactions_page(from_index: int, size: int) -> list[MempoolTransaction]:
    print(f"Getting transactions from index {from_index} to {from_index + size}")

    url = f"{MEMPOOL_URL}?from={from_index}&size={size}&type=Transaction"
    response = requests.get(url)
    records = response.json()
    transactions = [MempoolTransaction.from_record(record) for record in records]

    print(f"Got {len(transactions)} transactions")

    return transactions


def sort_transactions(transactions: list[MempoolTransaction]) -> list[MempoolTransaction]:
    transactions_heap: list[MempoolTransaction] = list(transactions)
    heapq.heapify(transactions_heap)

    sorted_transactions: list[MempoolTransaction] = [heapq.heappop(transactions_heap) for _ in range(len(transactions_heap))]

    for transaction in sorted_transactions:
        print(f"\t{transaction.sender}, nonce = {transaction.nonce}, ppu = {transaction.ppu:_}, gas_limit = {transaction.gas_limit:_}, gas_price = {transaction.gas_price:_}, fn = {transaction.function}")

    return sorted_transactions


def distribute_transactions_into_gas_buckets(transactions: list[MempoolTransaction]) -> list[GasBucket]:
    buckets: list[GasBucket] = []
    current_gas_bucket = GasBucket(index=0)

    for transaction in transactions:
        bucket_has_room = current_gas_bucket.add_transaction(transaction)

        if not bucket_has_room:
            buckets.append(current_gas_bucket)
            current_gas_bucket = GasBucket(index=current_gas_bucket.index + 1)

    # Don't forget to collect the last (could also be the first) bucket!
    buckets.append(current_gas_bucket)

    print("-" * 10)
    print("Number of buckets:", len(buckets))

    for bucket in buckets:
        print(f"Bucket {bucket.index}, gas = {bucket.gas_accumulated:_}, num_txs = {bucket.num_transactions}")
        print(f"\tppu: {bucket.ppu_begin:_} .. {bucket.ppu_end:_}")

    return buckets


if __name__ == "__main__":
    main()
