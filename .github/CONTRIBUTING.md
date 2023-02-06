
# Contributing to mx-chain-go

If you are unfamiliar with the workflow of contributing to GitHub, you can refer to [this article](https://github.com/firstcontributions/first-contributions/blob/master/README.md)

## External contributions 

If you are not part of the team and want to contribute to this repository, you must fork & clone this repository

The development should happen in a personal fork, cloned on the local machine. **Make sure you use signed commits before opening** a pull request.

For external contributors, the PRs should be targeted towards the `master` branch. A team responsible will instruct 
the PR owner to re-target it against another branch, in accordance to internal branches management.

**tl;dr**:
- fork `mx-chain-go` and use `master` branch.
- use signed commits. [docs](https://docs.github.com/en/authentication/managing-commit-signature-verification/signing-commits).
- open a PR against the `master` branch of `multiversx/mx-chain-go`.

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

## Branches management

Internal Branches/Releases Management (to be checked by both the code owner and the reviewers)

### `mx-chain-go`
If the PR is:
1. a hotfix: it will be targeted towards `master` branch.
2. a feature:
   2.1. a small feature (a single PR is needed): targeted towards an `rc` branch.
   2.2. a big feature (more than one PR is needed): create a feature branch (`feat/...`) that is targeted towards an `rc` branch.

### Satellite projects (`mx-chain-vm-common-go`, `mx-chain-core-go`, and so on)
If the PR is:
1. a hotfix: it will be targeted towards `master`/`main` branch.
2. a feature:
   2.1. a small feature (a single PR is needed): targeted towards an `rc` branch and create a new tag&release after merging, and reference it in `mx-chain-go` (if needed).
   2.2. a big feature:
   2.2.1. a small satellite PR: targeted towards an `rc` branch. For each change, reference just the commit hash on `mx-chain-go` and create a tag&pre-release when the `mx-chain-go` PR is to be merged.
   2.2.2. a big satellite PR: create a feature branch (`feat/...`) targeted towards an `rc` branch. For each change, reference just the commit hash on `mx-chain-go` and create a tag&pre-release when the `feat` branch is merged
