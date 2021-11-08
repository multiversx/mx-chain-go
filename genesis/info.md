# Index transactions

Implement, at first, for MintTransactions.
Parse file once for MintTransactions (tranzactii la care se ia suplyValue)
Parse file again for staking/delegation Transactions.

Create the transactions based on the args (also from the info in the ticket).
For each account create the transactions

Follow CreateShardGenesisBlock outer execution, and go to
factory/processComponents.go and check `indexGenesisBlocks`. That's the place
where the miniblocks (for each shard) has to be created, and added to the
indexer Body section. All the transactions  should be added to the coresponding
shard

Check `setBalancesToTrie`

- development flow
- index both transactions (separately) and miniblocks?
- create transactions again or use already existing flow?
- elastic indexer vs dblookupext index

Testing:
- `indexGenesisBlocks`
    - mock accounts parser
    - mock SaveBlock from outportHandler, and check that the arg has the corrent fields
