package keyshards

import (
	"crypto/md5"
	"crypto/rand"
	_ "crypto/rand"
	"io"

	"github.com/axengine/utils/crypto"
	"github.com/yalaorg/golibs/shamir"
)

var _shardEncryptKey_ = func() []byte {
	p1 := md5.Sum([]byte("4\nreal_size×3+witness_siz"))
	p2 := md5.Sum([]byte("var buf bytes.Buffer"))
	return append(p1[:], p2[:]...)
}
var ivLength = 16

type GetSecretKeyFn = func() []byte

// ShardAndEncrypt Shard and encrypt the key, and return the encrypted sharded HEX
func ShardAndEncrypt(key []byte, totalShares, threshold int, getSecretKey GetSecretKeyFn) ([][]byte, error) {
	byteParts, err := shamir.Split(key, totalShares, threshold)
	if err != nil {
		return nil, err
	}
	iv := make([]byte, ivLength)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	var encryptedShards [][]byte
	if getSecretKey == nil {
		getSecretKey = _shardEncryptKey_
	}
	for _, v := range byteParts {
		b, err := crypto.AES256CBCPKCS0Encrypt(v, iv, getSecretKey())
		if err != nil {
			return nil, err
		}
		plaint, err := encode(b, iv)
		if err != nil {
			return nil, err
		}
		encryptedShards = append(encryptedShards, plaint)
	}

	return encryptedShards, nil
}

// DecryptAndCombine Decrypt the shards and assemble them
func DecryptAndCombine(shards [][]byte, getSecretKey GetSecretKeyFn) ([]byte, error) {
	var decrypedShards [][]byte
	for _, v := range shards {

		cipher, iv := decode(v)
		if getSecretKey == nil {
			getSecretKey = _shardEncryptKey_
		}
		decrypedShard, err := crypto.AES256CBCPKCS0Decrypt(cipher, iv, getSecretKey())
		if err != nil {
			return nil, err
		}
		decrypedShards = append(decrypedShards, decrypedShard)
	}

	return shamir.Combine(decrypedShards)
}

func encode(cipher, iv []byte) ([]byte, error) {
	buf := make([]byte, len(cipher)+len(iv))
	copy(buf[:8], iv[:8])
	copy(buf[8:], cipher[:])
	copy(buf[len(buf)-8:], iv[8:])
	return buf, nil
}

func decode(plaint []byte) ([]byte, []byte) {
	iv := make([]byte, ivLength)
	copy(iv[:8], plaint[:8])
	copy(iv[8:], plaint[len(plaint)-8:])
	cipher := plaint[8 : len(plaint)-8]
	return cipher, iv
}
