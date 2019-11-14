package proto

// Header is the print friendly translation of the header fields
type Header struct {
	magic   string
	version byte
	length  uint32
}

// Magic is the magic field from the Header data
func (h Header) Magic() string {
	return h.magic
}

// Version is the version field from the Header data
func (h Header) Version() byte {
	return h.version
}

// Len is the record count from the Header data
func (h Header) Len() uint32 {
	return h.length
}
