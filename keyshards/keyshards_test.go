package keyshards

import (
	"encoding/hex"
	"testing"

	"github.com/axengine/yalagolibs/shamir"
)

func TestShamir(t *testing.T) {
	secretBuf := []byte("hello yala")
	parts := 9
	threshold := 5
	byteParts, err := shamir.Split(secretBuf, parts, threshold)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(byteParts))
	for _, v := range byteParts {
		t.Log(hex.EncodeToString(v))
	}

	var shards = [][]byte{}
	for i := 0; i < 5; i++ {
		shards = append(shards, byteParts[i])
	}

	key, err := shamir.Combine(shards)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(key))
}

func TestKeyShards(t *testing.T) {
	key := []byte("yala")
	shards, err := ShardAndEncrypt(key, 9, 5, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(shards))
	for _, v := range shards {
		t.Log(hex.EncodeToString(v))
	}

	recoverKey, err := DecryptAndCombine(shards[:5], nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(recoverKey))
}
