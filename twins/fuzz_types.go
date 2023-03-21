package twins

import (
	"fmt"

	"github.com/relab/hotstuff"
	"github.com/relab/hotstuff/internal/proto/hotstuffpb"
)

type FuzzMsg interface {
	ToMsg() any
	ToString(int) string
}

type ProposeMsg struct {
	hotstuff.ID
	hotstuffpb.Proposal
}

func (proposeFuzzMsg *ProposeMsg) ToMsg() any {
	proposeMsg := hotstuffpb.ProposalFromProto(&proposeFuzzMsg.Proposal)
	proposeMsg.ID = proposeFuzzMsg.ID
	return proposeMsg
}

func (proposeFuzzMsg *ProposeMsg) ToString(depth int) string {
	var tabs string = ""
	for i := 0; i < depth; i++ {
		tabs += "\t"
	}

	return fmt.Sprintf(
		"twins.ProposeMsg{\n"+
			"%s\tID: %v\n"+
			"%s\tProposal: %v\n"+
			"%s}",
		tabs, proposeFuzzMsg.ID,
		tabs, ProposalToString(&proposeFuzzMsg.Proposal, depth+1),
		tabs)
}

func ProposalToString(proposal *hotstuffpb.Proposal, depth int) string {

	var tabs string = ""
	for i := 0; i < depth; i++ {
		tabs += "\t"
	}

	return fmt.Sprintf(
		"hotstuffpb.Proposal{\n"+
			"%s\tBlock: %v\n"+
			"%s\tAggQC: %v\n"+
			"%s}",
		tabs, BlockToString(proposal.Block, depth+1),
		tabs, AggQCToString(proposal.AggQC, depth+1),
		tabs)
}

func BlockToString(block *hotstuffpb.Block, depth int) string {

	if block == nil {
		return "nil"
	}

	var tabs string = ""
	for i := 0; i < depth; i++ {
		tabs += "\t"
	}

	return fmt.Sprintf(
		"hotstuffpb.Block{\n"+
			"%s\tParent: %v\n"+
			"%s\tQC: %v\n"+
			"%s\tView: %v\n"+
			"%s\tCommand: %v\n"+
			"%s\tProposer: %v\n"+
			"%s}",
		tabs, block.Parent,
		tabs, QuorumCertToString(block.QC, depth+1),
		tabs, block.View,
		tabs, block.Command,
		tabs, block.Proposer,
		tabs)
}

type TimeoutMsg struct {
	hotstuff.ID
	hotstuffpb.TimeoutMsg
}

func (timeoutFuzzMsg *TimeoutMsg) ToMsg() any {
	timeoutMsg := hotstuffpb.TimeoutMsgFromProto(&timeoutFuzzMsg.TimeoutMsg)
	timeoutMsg.ID = timeoutFuzzMsg.ID
	return timeoutMsg
}

func (timeoutFuzzMsg *TimeoutMsg) ToString(depth int) string {

	var tabs string = ""
	for i := 0; i < depth; i++ {
		tabs += "\t"
	}

	return fmt.Sprintf(
		"twins.TimeoutMsg{\n"+
			"%s\tID: %v\n"+
			"%s\tTimeoutMsg: %v\n"+
			"%s}",
		tabs, timeoutFuzzMsg.ID,
		tabs, TimeoutMsgToString(&timeoutFuzzMsg.TimeoutMsg, depth+1),
		tabs)
}

func TimeoutMsgToString(timeoutMsg *hotstuffpb.TimeoutMsg, depth int) string {
	var tabs string = ""
	for i := 0; i < depth; i++ {
		tabs += "\t"
	}

	return fmt.Sprintf(
		"hotstuffpb.TimeoutMsg{\n"+
			"%s\tView: %v\n"+
			"%s\tSyncInfo: %v\n"+
			"%s\tViewSig: %v\n"+
			"%s\tMsgSig: %v\n"+
			"%s}",
		tabs, timeoutMsg.View,
		tabs, SyncInfoToString(timeoutMsg.SyncInfo, depth+1),
		tabs, QuorumSignatureToString(timeoutMsg.ViewSig, depth+1),
		tabs, QuorumSignatureToString(timeoutMsg.MsgSig, depth+1),
		tabs)
}

type VoteMsg struct {
	hotstuff.ID
	hotstuffpb.PartialCert
	Deferred bool
}

func (voteFuzzMsg *VoteMsg) ToMsg() any {
	voteMsg := hotstuff.VoteMsg{}
	voteMsg.PartialCert = hotstuffpb.PartialCertFromProto(&voteFuzzMsg.PartialCert)
	voteMsg.ID = voteFuzzMsg.ID
	voteMsg.Deferred = voteFuzzMsg.Deferred
	return voteMsg
}

func (VoteFuzzMsg *VoteMsg) ToString(depth int) string {

	var tabs string = ""
	for i := 0; i < depth; i++ {
		tabs += "\t"
	}

	return fmt.Sprintf(
		"twins.VoteMsg{\n"+
			"%s\tDeferred: %v\n"+
			"%s\tID: %v\n"+
			"%s\tPartialCert: %v\n"+
			"%s}",
		tabs, VoteFuzzMsg.ID,
		tabs, VoteFuzzMsg.Deferred,
		tabs, PartialCertToString(&VoteFuzzMsg.PartialCert, depth+1),
		tabs)
}

type NewViewMsg struct {
	hotstuff.ID
	hotstuffpb.SyncInfo
}

func (newViewFuzzMsg *NewViewMsg) ToMsg() any {
	newViewMsg := hotstuff.NewViewMsg{}
	newViewMsg.SyncInfo = hotstuffpb.SyncInfoFromProto(&newViewFuzzMsg.SyncInfo)
	newViewMsg.ID = newViewFuzzMsg.ID
	return newViewMsg
}

func (newViewFuzzMsg *NewViewMsg) ToString(depth int) string {
	var tabs string = ""
	for i := 0; i < depth; i++ {
		tabs += "\t"
	}

	return fmt.Sprintf(
		"twins.NewViewMsg{\n"+
			"%s\tID: %v\n"+
			"%s\tSyncInfo: %v\n"+
			"%s}",
		tabs, newViewFuzzMsg.ID,
		tabs, SyncInfoToString(&newViewFuzzMsg.SyncInfo, depth+1),
		tabs)
}

func SyncInfoToString(syncInfo *hotstuffpb.SyncInfo, depth int) string {

	if syncInfo == nil {
		return "nil"
	}

	var tabs string = ""
	for i := 0; i < depth; i++ {
		tabs += "\t"
	}

	return fmt.Sprintf(
		"hotstuffpb.SyncInfo{\n"+
			"%s\tQC: %v\n"+
			"%s\tTC: %v\n"+
			"%s\tAggQC: %v\n"+
			"%s}",
		tabs, QuorumCertToString(syncInfo.QC, depth+1),
		tabs, TimeoutCertToString(syncInfo.TC, depth+1),
		tabs, AggQCToString(syncInfo.AggQC, depth+1),
		tabs)
}

func QuorumCertToString(qc *hotstuffpb.QuorumCert, depth int) string {

	if qc == nil {
		return "nil"
	}

	var tabs string = ""
	for i := 0; i < depth; i++ {
		tabs += "\t"
	}

	return fmt.Sprintf(
		"hotstuffpb.QuorumCert{\n"+
			"%s\tSig: %v\n"+
			"%s\tView: %v\n"+
			"%s\tHash: %v\n"+
			"%s}",
		tabs, QuorumSignatureToString(qc.Sig, depth+1),
		tabs, qc.View,
		tabs, qc.Hash,
		tabs)
}

func TimeoutCertToString(tc *hotstuffpb.TimeoutCert, depth int) string {

	if tc == nil {
		return "nil"
	}

	var tabs string = ""
	for i := 0; i < depth; i++ {
		tabs += "\t"
	}

	return fmt.Sprintf(
		"hotstuffpb.TimeoutCert{\n"+
			"%s\tSig: %v\n"+
			"%s\tView: %v\n"+
			"%s}",
		tabs, QuorumSignatureToString(tc.Sig, depth+1),
		tabs, tc.View,
		tabs)
}

func AggQCToString(aggQC *hotstuffpb.AggQC, depth int) string {

	if aggQC == nil {
		return "nil"
	}

	var tabs string = ""
	for i := 0; i < depth; i++ {
		tabs += "\t"
	}

	QCsString := "map[\n"

	for key, value := range aggQC.QCs {
		QCsString += fmt.Sprintf("%v\t\t%v: %v\n", tabs, key, QuorumCertToString(value, depth+2))
	}

	QCsString += tabs + "\t]"

	return fmt.Sprintf(
		"hotstuffpb.AggQC{\n"+
			"%s\tQCs: %v\n"+
			"%s\tSig: %v\n"+
			"%s\tView: %v\n"+
			"%s}",
		tabs, QCsString,
		tabs, QuorumSignatureToString(aggQC.Sig, depth+1),
		tabs, aggQC.View,
		tabs)
}

func PartialCertToString(partialCert *hotstuffpb.PartialCert, depth int) string {
	var tabs string = ""
	for i := 0; i < depth; i++ {
		tabs += "\t"
	}

	return fmt.Sprintf(
		"hotstuffpb.PartialCert{\n"+
			"%s\tSig: %v\n"+
			"%s\tHash: %v\n"+
			"%s}",
		tabs, QuorumSignatureToString(partialCert.Sig, depth+1),
		tabs, partialCert.Hash,
		tabs)
}

func QuorumSignatureToString(quorumSignature *hotstuffpb.QuorumSignature, depth int) string {

	if quorumSignature == nil {
		return "nil"
	}

	var tabs string = ""
	for i := 0; i < depth; i++ {
		tabs += "\t"
	}

	sigString := ""
	if quorumSignature.Sig == nil {
		sigString = "nil"
	} else {
		switch quorumSignature.Sig.(type) {
		case *hotstuffpb.QuorumSignature_ECDSASigs:
			sigString = QuorumSignature_ECDSASigsToString(quorumSignature.Sig.(*hotstuffpb.QuorumSignature_ECDSASigs), depth+1)
		case *hotstuffpb.QuorumSignature_BLS12Sig:
			sigString = QuorumSignature_BLS12SigToString(quorumSignature.Sig.(*hotstuffpb.QuorumSignature_BLS12Sig), depth+1)
		}
	}

	return fmt.Sprintf(
		"hotstuffpb.QuorumSignature{\n"+
			"%s\tSig: %s\n"+
			"%s}",
		tabs, sigString,
		tabs)
}

func QuorumSignature_ECDSASigsToString(ECDASigs *hotstuffpb.QuorumSignature_ECDSASigs, depth int) string {

	if ECDASigs == nil {
		return "nil"
	}

	var tabs string = ""
	for i := 0; i < depth; i++ {
		tabs += "\t"
	}

	return fmt.Sprintf(
		"hotstuffpb.QuorumSignature_ECDSASigs{\n"+
			"%s\tECDASigs: %s\n"+
			"%s},",
		tabs, ECDSAMultiSignatureToString(ECDASigs.ECDSASigs, depth+1),
		tabs)
}

func ECDSAMultiSignatureToString(ECDASigs *hotstuffpb.ECDSAMultiSignature, depth int) string {

	if ECDASigs == nil {
		return "nil"
	}

	var tabs string = ""
	for i := 0; i < depth; i++ {
		tabs += "\t"
	}

	sigsString := "[\n"

	for _, sig := range ECDASigs.Sigs {
		sigsString += fmt.Sprintf("%s\t\t%s\n", tabs, ECDASigToString(sig, depth+2))
	}

	sigsString += tabs + "\t]"

	return fmt.Sprintf(
		"hotstuffpb.ECDSAMultiSignature{\n"+
			"%s\tECDASigs: %s\n"+
			"%s}",
		tabs, sigsString,
		tabs)
}

func ECDASigToString(sig *hotstuffpb.ECDSASignature, depth int) string {

	if sig == nil {
		return "nil"
	}

	var tabs string = ""
	for i := 0; i < depth; i++ {
		tabs += "\t"
	}

	return fmt.Sprintf(
		"hotstuffpb.ECDSASignature{\n"+
			"%s\tSigner: %v\n"+
			"%s\tR: %v\n"+
			"%s\tS: %v\n"+
			"%s}",
		tabs, sig.Signer,
		tabs, sig.R,
		tabs, sig.S,
		tabs)
}

func QuorumSignature_BLS12SigToString(BLS12Sig *hotstuffpb.QuorumSignature_BLS12Sig, depth int) string {

	if BLS12Sig == nil {
		return "nil"
	}

	var tabs string = ""
	for i := 0; i < depth; i++ {
		tabs += "\t"
	}

	return fmt.Sprintf(
		"hotstuffpb.QuorumSignature_BLS12Sig{\n"+
			"%s\tBLS12Sig: %s\n"+
			"%s},",
		tabs, BLS12AggregateSignature(BLS12Sig.BLS12Sig, depth+1),
		tabs)
}

func BLS12AggregateSignature(BLS12Sig *hotstuffpb.BLS12AggregateSignature, depth int) string {

	if BLS12Sig == nil {
		return "nil"
	}

	var tabs string = ""
	for i := 0; i < depth; i++ {
		tabs += "\t"
	}

	return fmt.Sprintf(
		"hotstuffpb.BLS12AggregateSignature{\n"+
			"%s\tSig: %v\n"+
			"%s\tParticipants: %v\n"+
			"%s},",
		tabs, BLS12Sig.Sig,
		tabs, BLS12Sig.Participants,
		tabs)
}
