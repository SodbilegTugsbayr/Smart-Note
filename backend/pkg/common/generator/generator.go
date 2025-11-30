package generator

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"strconv"
	"time"
)

// func main() {
// 	fmt.Printf("%s\n", RandomStr(60))
// 	fmt.Printf("%s\n", RandomSimpleStr(60))
// }

const letterBytes = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const symbolBytes = "!@#$^&*~<>"

func RandomString(n int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	a := make([]byte, n-n/3)
	for i := range a {
		a[i] = letterBytes[r.Intn(len(letterBytes))]
	}
	b := make([]byte, n/3)
	for i := range b {
		b[i] = symbolBytes[r.Intn(len(symbolBytes))]
	}
	chars := []byte(string(a) + string(b))
	r.Shuffle(len(chars), func(i, j int) {
		chars[i], chars[j] = chars[j], chars[i]
	})
	return string(chars)
}

func RandomSimpleString(n int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	a := make([]byte, n)
	for i := range a {
		a[i] = letterBytes[r.Intn(len(letterBytes))]
	}
	chars := []byte(string(a))
	r.Shuffle(len(chars), func(i, j int) {
		chars[i], chars[j] = chars[j], chars[i]
	})
	return string(chars)
}

func GenerateKey(blob []byte) ([]byte, error) {
	// Add timestamp to prevent key collision
	now := strconv.FormatInt(time.Now().UnixNano(), 10)
	h := md5.New()
	if _, err := h.Write(append(blob, []byte(now)...)); err != nil {
		return nil, err
	}
	// TODO: Test if we really need following line
	hash := hex.EncodeToString(h.Sum(nil))

	return []byte(hash), nil
}

func GenerateAPIKey(salt string) (string, error) {
	// Add timestamp to prevent key collision
	now := strconv.FormatInt(time.Now().UnixNano(), 10)
	h := sha256.New()
	if _, err := h.Write([]byte(RandomString(60))); err != nil {
		return "", err
	}
	if _, err := h.Write([]byte(now)); err != nil {
		return "", err
	}
	if _, err := h.Write([]byte(salt)); err != nil {
		return "", err
	}
	// TODO: Test if we really need following line
	hash := hex.EncodeToString(h.Sum(nil))

	return hash, nil
}

func GenerateNumbersInString(length int) string {
	s := ""
	for i := 0; i < length; i++ {
		s += (string)(rand.Intn(10) + 48)
	}
	return s
}
