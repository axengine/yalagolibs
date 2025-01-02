package bitcoinlib

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
)

func SignBitcoinMessage(sk *btcec.PrivateKey, msg []byte) []byte {
	var buf bytes.Buffer
	magicStr := "Bitcoin Signed Message:\n"
	buf.WriteByte(byte(len(magicStr)))
	buf.WriteString(magicStr)

	bz, err := encode(int64(len(msg)))
	if err != nil {
		panic(err)
	}

	buf.Write(bz)
	buf.Write(msg)

	hash := _sha256(_sha256(buf.Bytes()))

	return ecdsa.SignCompact(sk, hash, true)
}

func _sha256(bz []byte) []byte {
	h := sha256.New()
	h.Write(bz)
	sum := h.Sum(nil)
	return sum
}

const MAX_SAFE_INTEGER = 9007199254740991

// checkUInt53 checks if the number is a valid unsigned 53-bit integer.
func checkUInt53(n int64) error {
	if n < 0 || n > MAX_SAFE_INTEGER || n != n {
		return errors.New("value out of range")
	}
	return nil
}

// encodingLength returns the number of bytes needed to encode the number.
func encodingLength(number int64) int {
	if number < 0xfd {
		return 1
	} else if number <= 0xffff {
		return 3
	} else if number <= 0xffffffff {
		return 5
	}
	return 9
}

// encode encodes a number into a byte slice.
func encode(number int64) ([]byte, error) {
	if err := checkUInt53(number); err != nil {
		return nil, err
	}

	buffer := make([]byte, encodingLength(number))

	// 8 bit
	if number < 0xfd {
		buffer[0] = byte(number)
		return buffer, nil
		// 16 bit
	} else if number <= 0xffff {
		buffer[0] = 0xfd
		binary.LittleEndian.PutUint16(buffer[1:], uint16(number))
		return buffer, nil
		// 32 bit
	} else if number <= 0xffffffff {
		buffer[0] = 0xfe
		binary.LittleEndian.PutUint32(buffer[1:], uint32(number))
		return buffer, nil
		// 64 bit
	} else {
		buffer[0] = 0xff
		binary.LittleEndian.PutUint32(buffer[1:], uint32(number&0xffffffff))
		binary.LittleEndian.PutUint32(buffer[5:], uint32(number/0x100000000))
		return buffer, nil
	}
}
