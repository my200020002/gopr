package fuzhu

import (
	"crypto/md5"
	"encoding/hex"
	"hash/fnv"
)

// 计算数据的MD5哈希
func CalculateHashMD5(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}
func CalculateHashFNV(data []byte) uint64 {
	h := fnv.New64a()
	h.Write(data)
	return h.Sum64()
}
