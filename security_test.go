package ntlmssp

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignKey(t *testing.T) {
	tables := []struct {
		exportedSessionKey, constant, want []byte
	}{
		{
			bytes.Repeat([]byte{0x55}, 16),
			clientSigning,
			[]byte{
				0x47, 0x88, 0xdc, 0x86, 0x1b, 0x47, 0x82, 0xf3,
				0x5d, 0x43, 0xfd, 0x98, 0xfe, 0x1a, 0x2d, 0x39,
			},
		},
	}

	for _, table := range tables {
		got := signKey(table.exportedSessionKey, table.constant)
		assert.Equal(t, table.want, got)
	}
}

func TestSealKey(t *testing.T) {
	tables := []struct {
		flags                        uint32
		exportedSessionKey, constant []byte
		exchangeKey                  func(uint32, []byte) ([]byte, error)
		want                         []byte
	}{
		// No flags
		{
			0,
			bytes.Repeat([]byte{0x55}, 16),
			clientSealing,
			nil,
			bytes.Repeat([]byte{0x55}, 16),
		},
		// NTLMv1 56-bit
		{
			0x80000080,
			bytes.Repeat([]byte{0x55}, 16),
			clientSealing,
			nil,
			[]byte{
				0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0xa0,
			},
		},
		// NTLMv1 40-bit
		{
			0x80,
			bytes.Repeat([]byte{0x55}, 16),
			clientSealing,
			nil,
			[]byte{
				0x55, 0x55, 0x55, 0x55, 0x55, 0xe5, 0x38, 0xb0,
			},
		},
		// NTLMv2 128-bit
		{
			0x20080000,
			bytes.Repeat([]byte{0x55}, 16),
			clientSealing,
			nil,
			[]byte{
				0x59, 0xf6, 0x00, 0x97, 0x3c, 0xc4, 0x96, 0x0a,
				0x25, 0x48, 0x0a, 0x7c, 0x19, 0x6e, 0x4c, 0x58,
			},
		},
		// NTLMv2 56-bit
		{
			0x80080000,
			[]byte{
				0xd8, 0x72, 0x62, 0xb0, 0xcd, 0xe4, 0xb1, 0xcb,
				0x74, 0x99, 0xbe, 0xcc, 0xcd, 0xf1, 0x07, 0x84,
			},
			clientSealing,
			func(flags uint32, sessionBaseKey []byte) ([]byte, error) {
				return ntlmV1ExchangeKey(flags, sessionBaseKey, []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef}, zeroPad(bytes.Repeat([]byte{0xaa}, 8), 24), zeroPad(bytes.Repeat([]byte{0xaa}, 8), 16))
			},
			[]byte{
				0x04, 0xdd, 0x7f, 0x01, 0x4d, 0x85, 0x04, 0xd2,
				0x65, 0xa2, 0x5c, 0xc8, 0x6a, 0x3a, 0x7c, 0x06,
			},
		},
		// NTLMv2 40-bit
		{
			0x80000,
			[]byte{
				0xd8, 0x72, 0x62, 0xb0, 0xcd, 0xe4, 0xb1, 0xcb,
				0x74, 0x99, 0xbe, 0xcc, 0xcd, 0xf1, 0x07, 0x84,
			},
			clientSealing,
			func(flags uint32, sessionBaseKey []byte) ([]byte, error) {
				return ntlmV1ExchangeKey(flags, sessionBaseKey, []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef}, zeroPad(bytes.Repeat([]byte{0xaa}, 8), 24), zeroPad(bytes.Repeat([]byte{0xaa}, 8), 16))
			},
			[]byte{
				0x26, 0xb2, 0xc1, 0xe7, 0x7b, 0xe4, 0x53, 0x3d,
				0x55, 0x5a, 0x22, 0x0a, 0x0f, 0xde, 0xb9, 0x6c,
			},
		},
	}

	for _, table := range tables {
		var exportedSessionKey []byte
		if table.exchangeKey != nil {
			var err error
			exportedSessionKey, err = table.exchangeKey(table.flags, table.exportedSessionKey)
			assert.Nil(t, err)
		} else {
			exportedSessionKey = table.exportedSessionKey
		}

		got := sealKey(table.flags, exportedSessionKey, table.constant)
		assert.Equal(t, table.want, got)
	}
}

func TestSecuritySession(t *testing.T) {
	tables := []struct {
		flags              uint32
		exportedSessionKey []byte
		message            []byte
		seal, signature    []byte
		err                error
	}{
		{
			0xe2028233,
			bytes.Repeat([]byte{0x55}, 16),
			[]byte{
				0x50, 0x00, 0x6c, 0x00, 0x61, 0x00, 0x69, 0x00,
				0x6e, 0x00, 0x74, 0x00, 0x65, 0x00, 0x78, 0x00,
				0x74, 0x00,
			},
			[]byte{
				0x56, 0xfe, 0x04, 0xd8, 0x61, 0xf9, 0x31, 0x9a,
				0xf0, 0xd7, 0x23, 0x8a, 0x2e, 0x3b, 0x4d, 0x45,
				0x7f, 0xb8,
			},
			[]byte{
				0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x09, 0xdc, 0xd1, 0xdf, 0x2e, 0x45, 0x9d, 0x36,
			},
			nil,
		},
		{
			0x820a8233,
			[]byte{
				0xeb, 0x93, 0x42, 0x9a, 0x8b, 0xd9, 0x52, 0xf8,
				0xb8, 0x9c, 0x55, 0xb8, 0x7f, 0x47, 0x5e, 0xdc,
			},
			[]byte{
				0x50, 0x00, 0x6c, 0x00, 0x61, 0x00, 0x69, 0x00,
				0x6e, 0x00, 0x74, 0x00, 0x65, 0x00, 0x78, 0x00,
				0x74, 0x00,
			},
			[]byte{
				0xa0, 0x23, 0x72, 0xf6, 0x53, 0x02, 0x73, 0xf3,
				0xaa, 0x1e, 0xb9, 0x01, 0x90, 0xce, 0x52, 0x00,
				0xc9, 0x9d,
			},
			[]byte{
				0x01, 0x00, 0x00, 0x00, 0xff, 0x2a, 0xeb, 0x52,
				0xf6, 0x81, 0x79, 0x3a, 0x00, 0x00, 0x00, 0x00,
			},
			nil,
		},
		{
			0xe28a8234,
			bytes.Repeat([]byte{0x55}, 16),
			[]byte{
				0x50, 0x00, 0x6c, 0x00, 0x61, 0x00, 0x69, 0x00,
				0x6e, 0x00, 0x74, 0x00, 0x65, 0x00, 0x78, 0x00,
				0x74, 0x00,
			},
			[]byte{
				0x54, 0xe5, 0x01, 0x65, 0xbf, 0x19, 0x36, 0xdc,
				0x99, 0x60, 0x20, 0xc1, 0x81, 0x1b, 0x0f, 0x06,
				0xfb, 0x5f,
			},
			[]byte{
				0x01, 0x00, 0x00, 0x00, 0x7f, 0xb3, 0x8e, 0xc5,
				0xc5, 0x5d, 0x49, 0x76, 0x00, 0x00, 0x00, 0x00,
			},
			nil,
		},
	}

	for _, table := range tables {
		c, err := newSecuritySession(table.flags, table.exportedSessionKey, sourceClient)
		assert.Nil(t, err)
		s, err := newSecuritySession(table.flags, table.exportedSessionKey, sourceServer)
		assert.Nil(t, err)
		seal, signature, err := c.Wrap(table.message)
		assert.Equal(t, table.err, err)
		if err == nil {
			assert.Equal(t, table.seal, seal)
			assert.Equal(t, table.signature, signature)
		}
		message, err := s.Unwrap(seal, signature)
		assert.Equal(t, table.err, err)
		if err == nil {
			assert.Equal(t, table.message, message)
		}
	}
}
