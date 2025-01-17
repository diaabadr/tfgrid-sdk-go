// Package client provides a simple RMB interface to work with the node.
package client

import (
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/subi"
	"github.com/threefoldtech/tfgrid-sdk-go/rmb-sdk-go"
)

// NodeClientGetter is an interface for node client
type NodeClientGetter interface {
	GetNodeClient(sub subi.SubstrateExt, nodeID uint32) (*NodeClient, error)
}

// NodeClientPool is a pool for node clients and rmb
type NodeClientPool struct {
	nodeClients sync.Map
	rmb         rmb.Client
	timeout     time.Duration
}

// NewNodeClientPool generates a new client pool
func NewNodeClientPool(rmb rmb.Client, timeout time.Duration) *NodeClientPool {
	return &NodeClientPool{
		nodeClients: sync.Map{},
		rmb:         rmb,
		timeout:     timeout,
	}
}

// GetNodeClient gets the node client according to node ID
func (p *NodeClientPool) GetNodeClient(sub subi.SubstrateExt, nodeID uint32) (*NodeClient, error) {
	cl, ok := p.nodeClients.Load(nodeID)
	if ok {
		return cl.(*NodeClient), nil
	}

	twinID, err := sub.GetNodeTwin(nodeID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get node %d", nodeID)
	}

	cl, _ = p.nodeClients.LoadOrStore(nodeID, NewNodeClient(uint32(twinID), p.rmb, p.timeout))

	return cl.(*NodeClient), nil
}
