package ecdsa

import (
	"container/list"
	"sync"
	"sync/atomic"

	"github.com/relab/hotstuff"
)

type signatureCache struct {
	ecdsaCrypto

	mut      sync.Mutex
	cache    map[string]*list.Element
	capacity int
	order    list.List
}

// NewWithCache returns a new signer and a new verifier that use caching to speed up verification
func NewWithCache(capacity int) (hotstuff.Signer, hotstuff.Verifier) {
	cache := &signatureCache{
		ecdsaCrypto: ecdsaCrypto{},
		cache:       make(map[string]*list.Element, capacity),
		capacity:    capacity,
	}
	return cache, cache
}

func (c *signatureCache) dropOldest() {
	elem := c.order.Back()
	delete(c.cache, elem.Value.(string))
	c.order.Remove(elem)
}

func (c *signatureCache) insert(key string) {
	if elem, ok := c.cache[key]; ok {
		c.order.MoveToFront(elem)
		return
	}

	if len(c.cache)+1 > c.capacity {
		c.dropOldest()
	}

	elem := c.order.PushFront(key)
	c.cache[key] = elem
}

func (c *signatureCache) check(key string) bool {
	elem, ok := c.cache[key]
	if !ok {
		return false
	}
	c.order.MoveToFront(elem)
	return true
}

// Sign signs a hash.
func (c *signatureCache) Sign(hash hotstuff.Hash) (sig hotstuff.Signature, err error) {
	sig, err = c.ecdsaCrypto.Sign(hash)
	if err != nil {
		return nil, err
	}
	k := string(sig.ToBytes())
	c.mut.Lock()
	c.insert(k)
	c.mut.Unlock()
	return sig, nil
}

// CreatePartialCert signs a single block and returns the partial certificate.
func (c *signatureCache) CreatePartialCert(block *hotstuff.Block) (cert hotstuff.PartialCert, err error) {
	signature, err := c.ecdsaCrypto.CreatePartialCert(block)
	if err != nil {
		return nil, err
	}
	k := string(signature.ToBytes())
	c.mut.Lock()
	c.insert(k)
	c.mut.Unlock()
	return signature, nil
}

// Verify verifies a signature given a hash.
func (c *signatureCache) Verify(sig hotstuff.Signature, hash hotstuff.Hash) bool {
	k := string(sig.ToBytes())
	c.mut.Lock()
	if c.check(k) {
		c.mut.Unlock()
		return true
	}
	c.mut.Unlock()

	if !c.ecdsaCrypto.Verify(sig, hash) {
		return false
	}

	c.mut.Lock()
	c.insert(k)
	c.mut.Unlock()

	return true
}

// VerifyPartialCert verifies a single partial certificate.
func (c *signatureCache) VerifyPartialCert(cert hotstuff.PartialCert) bool {
	k := string(cert.Signature().ToBytes())

	c.mut.Lock()
	if c.check(k) {
		c.mut.Unlock()
		return true
	}
	c.mut.Unlock()

	if !c.ecdsaCrypto.VerifyPartialCert(cert) {
		return false
	}

	c.mut.Lock()
	c.insert(k)
	c.mut.Unlock()

	return true
}

func (c *signatureCache) verifyAggregateSignature(agg aggregateSignature, hash hotstuff.Hash) bool {
	if len(agg) < c.ecdsaCrypto.mod.Config().QuorumSize() {
		return false
	}

	var wg sync.WaitGroup
	var numValid uint32

	// first check if any signatures are cache
	c.mut.Lock()
	for _, sig := range agg {
		k := string(sig.ToBytes())
		if c.check(k) {
			numValid++
			continue
		}
		// on cache miss, we start a goroutine to verify the signature
		wg.Add(1)
		go func(sig *Signature) {
			if c.ecdsaCrypto.Verify(sig, hash) {
				atomic.AddUint32(&numValid, 1)
				c.mut.Lock()
				c.insert(string(sig.ToBytes()))
				c.mut.Unlock()
			}
			wg.Done()
		}(sig)
	}
	c.mut.Unlock()

	wg.Wait()
	return numValid >= uint32(c.ecdsaCrypto.mod.Config().QuorumSize())
}

// VerifyQuorumCert verifies a quorum certificate.
func (c *signatureCache) VerifyQuorumCert(qc hotstuff.QuorumCert) bool {
	// If QC was created for genesis, then skip verification.
	if qc.BlockHash() == hotstuff.GetGenesis().Hash() {
		return true
	}

	ecdsaQC := qc.(*QuorumCert)
	return c.verifyAggregateSignature(ecdsaQC.signatures, ecdsaQC.hash)
}

// VerifyTimeoutCert verifies a timeout certificate.
func (c *signatureCache) VerifyTimeoutCert(tc hotstuff.TimeoutCert) bool {
	ecdsaTC := tc.(*TimeoutCert)
	return c.verifyAggregateSignature(ecdsaTC.signatures, ecdsaTC.view.ToHash())
}