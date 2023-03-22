package twins

import (
	"github.com/relab/hotstuff"
	"github.com/relab/hotstuff/internal/proto/hotstuffpb"
)

type FuzzMsg interface {
	ToMsg() any
	ToString(int) string
	String() string
}

type ProposeMsg struct {
	hotstuff.ID
	hotstuffpb.Proposal
}

func (fuzzMsg *ProposeMsg) String() string {
	return fuzzMsg.ToString(0)
}

func (proposeFuzzMsg *ProposeMsg) ToMsg() any {
	proposeMsg := hotstuffpb.ProposalFromProto(&proposeFuzzMsg.Proposal)
	proposeMsg.ID = proposeFuzzMsg.ID
	return proposeMsg
}

type TimeoutMsg struct {
	hotstuff.ID
	hotstuffpb.TimeoutMsg
}

func (fuzzMsg *TimeoutMsg) String() string {
	return fuzzMsg.ToString(0)
}

func (timeoutFuzzMsg *TimeoutMsg) ToMsg() any {
	timeoutMsg := hotstuffpb.TimeoutMsgFromProto(&timeoutFuzzMsg.TimeoutMsg)
	timeoutMsg.ID = timeoutFuzzMsg.ID
	return timeoutMsg
}

type VoteMsg struct {
	hotstuff.ID
	hotstuffpb.PartialCert
	Deferred bool
}

func (fuzzMsg *VoteMsg) String() string {
	return fuzzMsg.ToString(0)
}

func (voteFuzzMsg *VoteMsg) ToMsg() any {
	voteMsg := hotstuff.VoteMsg{}
	voteMsg.PartialCert = hotstuffpb.PartialCertFromProto(&voteFuzzMsg.PartialCert)
	voteMsg.ID = voteFuzzMsg.ID
	voteMsg.Deferred = voteFuzzMsg.Deferred
	return voteMsg
}

type NewViewMsg struct {
	hotstuff.ID
	hotstuffpb.SyncInfo
}

func (fuzzMsg *NewViewMsg) String() string {
	return fuzzMsg.ToString(0)
}

func (newViewFuzzMsg *NewViewMsg) ToMsg() any {
	newViewMsg := hotstuff.NewViewMsg{}
	newViewMsg.SyncInfo = hotstuffpb.SyncInfoFromProto(&newViewFuzzMsg.SyncInfo)
	newViewMsg.ID = newViewFuzzMsg.ID
	return newViewMsg
}
