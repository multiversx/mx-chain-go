## Reasoning behind the pull request
- 
- 
- 
  
## Proposed changes
- 
- 
- 

## Testing procedure
- 
- 
- 

## Pre-requisites

Based on the [Contributing Guidelines](https://github.com/multiversx/mx-chain-go/blob/master/.github/CONTRIBUTING.md#branches-management) the PR author and the reviewers must check the following requirements are met:
- was the PR targeted to the correct branch?
- if this is a larger feature that probably needs more than one PR, is there a `feat` branch created?
- if this is a `feat` branch merging, do all satellite projects have a proper tag inside `go.mod`?

## Triggering Tests from Comments
To re-run integration tests with custom branches, add a comment in this format:

```
Run Tests:
mx-chain-simulator-go: feat-sub-second-round-2
mx-chain-testing-suite: fix_ci
```