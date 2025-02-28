package model

import (
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

type AxelarAmplifierVerifierVote struct {
	bun.BaseModel `bun:"table:axelar_amplifier_verifier"`

	ID              int64              `bun:"id,pk,autoincrement"`
	ChainInfoID     int64              `bun:"chain_info_id,pk,notnull"`
	CreatedAt       time.Time          `bun:"created_at"`
	ChainAndPollID  string             `bun:"chain_and_poll_id"`
	PollStartHeight int64              `bun:"poll_start_height"`
	PollVoteHeight  int64              `bun:"poll_vote_height"`
	VerifierID      int64              `bun:"verifier_id"`
	Status          VerifierVoteStatus `bun:"status"`
}

func (model *AxelarAmplifierVerifierVote) UpdateVote(statusStr string, voteHeight int64) {
	model.Status = StringToPollVote(statusStr)
	model.PollVoteHeight = voteHeight
}

// String method for human-readable output
func (v AxelarAmplifierVerifierVote) String() string {
	return fmt.Sprintf("Vote(PollID: %s, PollStartHeight=%d, VerifierAddressID=%d, Status=%d", v.ChainAndPollID, v.PollStartHeight, v.VerifierID, v.Status)
}

/*
	pub enum Vote {
	    SucceededOnChain, // the txn was included on chain, and achieved the intended result
	    FailedOnChain,    // the txn was included on chain, but failed to achieve the intended result
	    NotFound,         // the txn could not be found on chain in any blocks at the time of voting
	}
*/
type VerifierVoteStatus int64

var (
	PollStart        VerifierVoteStatus = 0
	FailedOnChain    VerifierVoteStatus = 1
	NotFound         VerifierVoteStatus = 2
	SucceededOnChain VerifierVoteStatus = 3
)

func StringToPollVote(str string) VerifierVoteStatus {
	switch str {
	case "failed_on_chain":
		return FailedOnChain
	case "not_found":
		return NotFound
	case "succeeded_on_chain":
		return SucceededOnChain
	}
	return PollStart
}

func (v VerifierVoteStatus) ToString() string {
	switch v {
	case FailedOnChain:
		return "failed_on_chain"
	case NotFound:
		return "not_found"
	case SucceededOnChain:
		return "succeeded_on_chain"
	}
	return "did_not_vote"
}

type RecentVote struct {
	Moniker string `bun:"moniker"`

	DidNotVote      int64  `bun:"did_not_vote"`
	DidNotVotePolls string `bun:"did_not_vote_polls"`

	FailedOnChain      int64  `bun:"failed_on_chain"`
	FailedOnChainPolls string `bun:"failed_on_chain_polls"`

	NotFound      int64  `bun:"not_found"`
	NotFoundPolls string `bun:"not_found_polls"`

	SucceededOnChain      int64  `bun:"succeeded_on_chain"`
	SucceededOnChainPolls string `bun:"succeeded_on_chain_polls"`
}
