package pelle2e

import "fmt"

type KVStoreAppContext struct {
	Key   string
	Value string
	Data  []byte
}

func NewKVStoreAppContext(key, value string) KVStoreAppContext {
	kvc := KVStoreAppContext{
		Key:   key,
		Value: value,
		Data:  []byte(fmt.Sprintf("%s=%s", key, value)),
	}
	return kvc
}

type E2EContext struct {
	KVStoreApp KVStoreAppContext
}

func NewE2EContext(kvctx KVStoreAppContext) *E2EContext {
	eectx := &E2EContext{
		KVStoreApp: kvctx,
	}
	return eectx
}

func (kvc *KVStoreAppContext) GenResponseDigest() [32]byte {
	var result [32]byte
	var valueByptes = []byte(kvc.Value)
	copy(result[:], valueByptes)
	return result
}
