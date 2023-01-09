package internal

import "errors"

var (
	errNoBlocks              = errors.New("No blocks found")
	errNoConnectedNodesFound = errors.New("No connected nodes found")
	errNoRndMsgsFound        = errors.New("no rnd messages found")
	errNoTxToFromBlock       = errors.New("No tx messages found to from block")
	errNotAValidator         = errors.New("Node not a viladator")
)
