package util

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// ReadTrustedKeys reads list of trusted public keys from file (server).
func ReadTrustedKeys(home string) (map[[32]byte]struct{}, error) {

	kd := filepath.Join(home, ".gclpr")
	_, err := os.Stat(kd)
	if err != nil {
		return nil, err
	}

	fn := filepath.Join(kd, "trusted")
	err = checkPermissions(fn, true)
	if err != nil {
		return nil, fmt.Errorf("trusted keys file permissions are too open: %w", err)
	}
	content, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, fmt.Errorf("unable to read public key: %w", err)
	}

	res := make(map[[32]byte]struct{})
	for _, b := range bytes.Split(bytes.ReplaceAll(content, []byte{'\r'}, []byte{'\n'}), []byte{'\n'}) {
		b = bytes.TrimSpace(b)
		if len(b) == 0 || b[0] == '#' {
			continue
		}
		l := hex.DecodedLen(len(b))
		if l != 32 {
			log.Printf("Wrong size for key %s... in trusted keys file. Ignoring\n", string(b[:min(8, l)]))
			continue
		}
		dst := make([]byte, l)
		n, err := hex.Decode(dst, b)
		if err != nil {
			log.Printf("Bad key %s... in trusted keys file: %s. Ignoring\n", string(b[:min(8, l)]), err.Error())
			continue
		}
		if n != 32 {
			log.Printf("Wrong size for key %s... in trusted keys file. Ignoring\n", string(b[:min(8, l)]))
			continue
		}
		var k [32]byte
		copy(k[:], dst)
		if _, ok := res[k]; ok {
			log.Printf("Duplicate key %s... in trusted keys file. Ignoring\n", string(b[:8]))
		}
		res[k] = struct{}{}
	}
	return res, nil
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
