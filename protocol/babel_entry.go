package protocol

// See RFC 8966 "The Babel Routing Protocol"
type BabelEntry struct {
	Prefix                string
	RouterId              string
	Metric                int64
	SequenceNumber        uint16
	Routes                int64
	Sources		      int64
}
