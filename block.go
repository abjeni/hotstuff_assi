package hotstuff

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

type Block struct {
	// keep a copy of the hash to avoid hashing multiple times
	hash     *Hash
	parent   Hash
	proposer ID
	cmd      Command
	cert     QuorumCert
	view     View
}

func NewBlock(parent Hash, cert QuorumCert, cmd Command, view View, proposer ID) *Block {
	return &Block{
		parent:   parent,
		cert:     cert,
		cmd:      cmd,
		view:     view,
		proposer: proposer,
	}
}

func (b *Block) String() string {
	return fmt.Sprintf(
		"Block{ hash: %.6s parent: %.6s, proposer: %d, view: %d , cert: %v }",
		b.Hash().String(),
		b.parent.String(),
		b.proposer,
		b.view,
		b.cert,
	)
}

func (b *Block) hashSlow() Hash {
	return sha256.Sum256(b.ToBytes())
}

// Hash returns the hash of theBlock
func (b *Block) Hash() Hash {
	if b.hash == nil {
		b.hash = new(Hash)
		*b.hash = b.hashSlow()
	}
	return *b.hash
}

// Proposer returns the id of the proposer
func (b *Block) Proposer() ID {
	return b.proposer
}

// Parent returns the hash of the parentBlock
func (b *Block) Parent() Hash {
	return b.parent
}

// Command returns the command
func (b *Block) Command() Command {
	return b.cmd
}

// Certificate returns the certificate that this Block references
func (b *Block) QuorumCert() QuorumCert {
	return b.cert
}

// View returns the view in which the Block was proposed
func (b *Block) View() View {
	return b.view
}

func (b *Block) ToBytes() []byte {
	buf := b.parent[:]
	var proposerBuf [4]byte
	binary.LittleEndian.PutUint32(proposerBuf[:], uint32(b.proposer))
	buf = append(buf, proposerBuf[:]...)
	var viewBuf [8]byte
	binary.LittleEndian.PutUint64(viewBuf[:], uint64(b.view))
	buf = append(buf, viewBuf[:]...)
	buf = append(buf, []byte(b.cmd)...)
	// genesis and dummy nodes have no certificates
	if b.cert != nil {
		buf = append(buf, b.cert.ToBytes()...)
	}
	return buf
}
