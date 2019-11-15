package proto

// Header is the protocol header data
type Header struct {
	prefix  string
	version byte
	length  uint32
}

// Prefix is the magic field from the header data
func (h Header) Prefix() string {
	return h.prefix
}

// Version is the version field from the header data
func (h Header) Version() byte {
	return h.version
}

// Len is the record count from the header data
func (h Header) Len() uint32 {
	return h.length
}

// Protocol is required for the ProtData interface
func (h Header) Protocol() {
	//todo indicate MPS7 support
}
