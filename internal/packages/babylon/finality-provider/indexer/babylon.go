package indexer

type fpVoteMap map[string]int64

type FinalityVoteSummary struct {
	BlockHeight           int64
	FinalityProviderVotes fpVoteMap
}
