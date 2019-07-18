package ntlmssp

import (
	"bytes"
	"encoding/binary"
)

// Adapted from golang.org/x/sys/windows

type filetime struct {
	LowDateTime  uint32
	HighDateTime uint32
}

func nsecToFiletime(nsec int64) (ft filetime) {
	// convert into 100-nanosecond
	nsec /= 100
	// change starting time to January 1, 1601
	nsec += 116444736000000000
	// split into high / low
	ft.LowDateTime = uint32(nsec & 0xffffffff)
	ft.HighDateTime = uint32(nsec >> 32 & 0xffffffff)
	return ft
}

func (ft *filetime) Marshal() ([]byte, error) {
	b := bytes.Buffer{}
	if err := binary.Write(&b, binary.LittleEndian, ft); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
