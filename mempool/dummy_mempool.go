package mempool

import (
	"time"

	"golang.org/x/exp/rand"
)

const numTxs = 100 // kb

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(rand.NewSource(uint64(time.Now().UnixNano())))

func generateRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
func ProduceTx() [][]byte {
	txBytes := make([][]byte, numTxs)
	for i := 0; i < numTxs; i++ {
		txString1 := generateRandomString(2)
		txString2 := generateRandomString(5)
		txString := txString1 + "=" + txString2
		txBytes[i] = []byte(txString)

	}
	return txBytes
}
