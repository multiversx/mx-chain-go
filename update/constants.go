package update

import "time"

// TimeoutGettingTrieNodes represents the maximum time allowed between 2 nodes fetches (and commits)
const TimeoutGettingTrieNodes = time.Minute * 10
