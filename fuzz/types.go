package fuzz

import (
	"github.com/relab/hotstuff"
	"github.com/relab/hotstuff/internal/proto/hotstuffpb"
)

type FuzzMsgInt interface {
	ToMsg() any
	ToString(int) string
	String() string
}

/*type ProposeMsg struct {
	hotstuff.ID
	hotstuffpb.Proposal
}*/

/*func (fuzzMsg *ProposeMsg) String() string {
	return fuzzMsg.ToString(0)
}*/

type AlmostFuzzMsg interface {
	Msg() FuzzMsgInt
}

func (proposeFuzzMsg *FuzzMsg_ProposeMsg) Msg() FuzzMsgInt {
	return proposeFuzzMsg.ProposeMsg
}

func (timeoutFuzzMsg *FuzzMsg_TimeoutMsg) Msg() FuzzMsgInt {
	return timeoutFuzzMsg.TimeoutMsg
}

func (newViewFuzzMsg *FuzzMsg_NewViewMsg) Msg() FuzzMsgInt {
	return newViewFuzzMsg.NewViewMsg
}

func (voteFuzzMsg *FuzzMsg_VoteMsg) Msg() FuzzMsgInt {
	return voteFuzzMsg.VoteMsg
}

func (fuzzMsg *FuzzMsg) Msg() FuzzMsgInt {
	return fuzzMsg.Message.(AlmostFuzzMsg).Msg()
}

func (proposeFuzzMsg *ProposeMsg) ToMsg() any {
	proposeMsg := hotstuffpb.ProposalFromProto(proposeFuzzMsg.Proposal)
	proposeMsg.ID = hotstuff.ID(proposeFuzzMsg.ID)
	return proposeMsg
}

/*type TimeoutMsg struct {
	hotstuff.ID
	hotstuffpb.TimeoutMsg
}*/

/*func (fuzzMsg *TimeoutMsg) String() string {
	return fuzzMsg.ToString(0)
}*/

func (timeoutFuzzMsg *TimeoutMsg) ToMsg() any {
	timeoutMsg := hotstuffpb.TimeoutMsgFromProto(timeoutFuzzMsg.TimeoutMsg)
	timeoutMsg.ID = hotstuff.ID(timeoutFuzzMsg.ID)
	return timeoutMsg
}

/*type VoteMsg struct {
	hotstuff.ID
	hotstuffpb.PartialCert
	Deferred bool
}*/

/*func (fuzzMsg *VoteMsg) String() string {
	return fuzzMsg.ToString(0)
}*/

func (voteFuzzMsg *VoteMsg) ToMsg() any {
	voteMsg := hotstuff.VoteMsg{}
	voteMsg.PartialCert = hotstuffpb.PartialCertFromProto(voteFuzzMsg.PartialCert)
	voteMsg.ID = hotstuff.ID(voteFuzzMsg.ID)
	voteMsg.Deferred = voteFuzzMsg.Deferred
	return voteMsg
}

/*type NewViewMsg struct {
	hotstuff.ID
	hotstuffpb.SyncInfo
}*/

/*func (fuzzMsg *NewViewMsg) String() string {
	return fuzzMsg.ToString(0)
}*/

func (newViewFuzzMsg *NewViewMsg) ToMsg() any {
	newViewMsg := hotstuff.NewViewMsg{}
	newViewMsg.SyncInfo = hotstuffpb.SyncInfoFromProto(newViewFuzzMsg.SyncInfo)
	newViewMsg.ID = hotstuff.ID(newViewFuzzMsg.ID)
	return newViewMsg
}
