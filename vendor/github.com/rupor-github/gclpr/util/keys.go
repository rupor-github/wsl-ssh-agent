package util

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/crypto/nacl/sign"
)

// ReadKeys returns previously generated key pair (client).
func ReadKeys(home string) (*[32]byte, *[64]byte, error) {

	kd := filepath.Join(home, ".gclpr")
	fi, err := os.Stat(kd)
	if err == nil && !fi.IsDir() {
		return nil, nil, fmt.Errorf("%s exists and is not a directory", kd)
	}
	if err != nil {
		return nil, nil, err
	}

	fn := filepath.Join(kd, "key.pub")
	err = checkPermissions(fn, true)
	if err != nil {
		return nil, nil, fmt.Errorf("public key file permissions are too open: %w", err)
	}
	pubkey, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to read public key: %w", err)
	}
	if len(pubkey) != 32 {
		return nil, nil, fmt.Errorf("bad public key size %d", len(pubkey))
	}

	fn = filepath.Join(kd, "key")
	err = checkPermissions(fn, false)
	if err != nil {
		return nil, nil, fmt.Errorf("private key file permissions are too open: %w", err)
	}
	key, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to read private key: %w", err)
	}
	if len(key) != 64 {
		return nil, nil, fmt.Errorf("bad private key size %d", len(key))
	}

	var pk [32]byte
	copy(pk[:], pubkey)
	var k [64]byte
	copy(k[:], key)
	return &pk, &k, nil
}

// CreateKeys generates and saves new keypair. If one exists - it will be overwritten (client).
func CreateKeys(home string) (*[32]byte, *[64]byte, error) {

	kd := filepath.Join(home, ".gclpr")
	if err := os.MkdirAll(kd, 0700); err != nil {
		return nil, nil, fmt.Errorf("cannot create keys directory %s: %w", kd, err)
	}

	pk, k, err := sign.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot generate keys: %w", err)
	}

	//nolint:gosec
	err = ioutil.WriteFile(filepath.Join(kd, "key.pub"), pk[:], 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to save public key: %w", err)
	}

	err = ioutil.WriteFile(filepath.Join(kd, "key"), k[:], 0600)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to save private key: %w", err)
	}
	return pk, k, nil
}

// ReadTrustedKeys reads list of trusted public keys from file (server).
func ReadTrustedKeys(home string) (map[[32]byte][32]byte, error) {

	kd := filepath.Join(home, ".gclpr")
	fi, err := os.Stat(kd)
	if err == nil && !fi.IsDir() {
		return nil, fmt.Errorf("%s exists and is not a directory", kd)
	}
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

	res := make(map[[32]byte][32]byte)
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

		var k [32]byte // public key
		copy(k[:], dst)
		hk := sha256.Sum256(dst) // and its hash
		if _, ok := res[hk]; ok {
			log.Printf("Duplicate key %s... in trusted keys file. Ignoring\n", string(b[:8]))
		}
		res[hk] = k
	}
	return res, nil
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ----------------------------------------------------------------------------
// Rest of the code calculates GnuPG compatible keygrip for ed25519 public key
// We may want to sent it out instead of public key itself one day.
//
// Ref: https://github.com/romanz/trezor-agent/blob/master/libagent/gpg/protocol.py
// ----------------------------------------------------------------------------

// bignum2bytes converts an unsigned integer to MSB-first byte slice with specified size.
func bignum2bytes(num string, size int) []byte {
	data, err := hex.DecodeString(num)
	if err != nil {
		log.Printf("Unable to decode [%s]: %s", num, err.Error())
		return nil
	}
	buf := make([]byte, size)
	for i := range buf {
		buf[i] = 0
	}
	for i, j := max(0, size-len(data)), 0; i < size; i, j = i+1, j+1 {
		buf[i] = data[j]

	}
	return buf
}

type part struct {
	name  string
	value []byte
}

func compute(parts []part) []byte {
	h := new(bytes.Buffer)
	for i := 0; i < len(parts); i++ {
		// buuld Rivest's S-exp
		if _, err := fmt.Fprintf(h, "(%d:%s%d:%s)", len(parts[i].name), parts[i].name, len(parts[i].value), parts[i].value); err != nil {
			log.Printf("IO error in keygrip compute: %s", err.Error())
			return nil
		}
	}
	s := sha1.Sum(h.Bytes())
	return s[:]
}

// GPGKeyGripED25519 computes GPG keygrip for Ed25519 public keys.
func GPGKeyGripED25519(pk [32]byte) []byte {
	return compute(
		[]part{
			{name: "p", value: bignum2bytes(ed25519_p, 32)},
			{name: "a", value: []byte{1}},
			{name: "b", value: bignum2bytes(ed25519_b, 32)},
			{name: "g", value: bignum2bytes(ed25519_g, 65)},
			{name: "n", value: bignum2bytes(ed25519_n, 32)},
			{name: "q", value: pk[:]},
		},
	)
}
