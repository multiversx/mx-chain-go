
# Contributing to elrond-go

If you are unfamiliar with the workflow of contributing to github, you can refer to this [this article](https://github.com/firstcontributions/first-contributions/blob/master/README.md)

## Fork & clone this repository

The development should happen in a personal fork, cloned on the local machine.

## Use development branch

External contributions should happen against the `development` branch.

Other branches may be used by the team members in accordance to internal decisions.

## Use linter

Make sure the code is well-formatted and aligned before opening a PR. Some linters such `gofmt` can be used, or code style
can be proactively checked while writing code, when using an IDE.

## Writing tests

In case of adding a new feature, make sure you follow these guidelines:
- write unit tests which verify the component in isolation and mock the external components around it;
- if the component relies on many other external components which are difficult to mock, write integrations tests for it; Although, this usually means that further decoupling using interfaces is required. Also, the component might needed to be split in smaller, specialized sub-components;
- cover both expected functionality but also verify that the code fails in expected ways.

In case of fixing a bug:
- write the unit test that exposes the bug, then fix it and have a test that will make sure the specific situation will always be tested in the future.

In both cases, make sure that the new code has a good code coverage (ideally, 100% of the new lines are covered) and also try to cover edge-cases.

Please take care of the severity of the bug. Pushing directly a fix of an undiscovered, critical bug, or a bug that can be exploited in such a way that could create damage to the blockchain, loss of funds, state alteration beyond repair, and so on, can be exploited before the team can take any action. Please contact the team on private channels before pushing the fix if you have any doubts.

## Make sure the tests pass

Before opening a Pull Request, make sure that all tests pass. One can run the following commands before opening a pull request:
- `make test`

## Manual testing

Although the nominal use case looks good, the linter runs without issues, as well as the unit & integration tests pass without any error, some unforeseen bugs may arise.

Depending on the PR type, there are multiple ways of manual testing the new code:
- if it only affects an API, a log, or something not critical, one could sync a node with new code on Testnet or Devnet and see if everything works good.
- if the changes would affect the entire network (consensus, processing, and so on), a local testnet should be started. Also, make sure backwards compatibility is maintained.
