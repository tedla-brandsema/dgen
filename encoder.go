package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

type CharGroup struct {
	Name     string
	Chars    string
	Required bool
}

type CharSet struct {
	Groups []CharGroup
}

func DefaultCharSet() *CharSet {
	return &CharSet{
		Groups: []CharGroup{
			{"lowercase", "abcdefghijkmnopqrstuvwxyz", true},
			{"uppercase", "ABCDEFGHJKLMNPQRSTUVWXYZ", true},
			{"digits", "123456789", true},
			{"special", "!@#$%^&*()-_=+[]{};:,.<>/?", true},
		},
	}
}

// SimplePRNG is a deterministic pseudorandom number generator using a xorshift32 algorithm.
// This implementation is explicitly defined to ensure identical results on any platform.
type SimplePRNG struct {
	state uint32
}

// NewSimplePRNG creates a new SimplePRNG with a non-zero seed.
func NewSimplePRNG(seed uint32) *SimplePRNG {
	if seed == 0 {
		seed = 1
	}
	return &SimplePRNG{state: seed}
}

// next returns the next pseudorandom uint32 using the xorshift32 algorithm.
func (prng *SimplePRNG) next() uint32 {
	prng.state ^= prng.state << 13
	prng.state ^= prng.state >> 17
	prng.state ^= prng.state << 5
	return prng.state
}

// Intn returns a pseudorandom integer in the range [0, n).
// This method is deterministic and defined by our own PRNG.
func (prng *SimplePRNG) Intn(n int) int {
	return int(prng.next() % uint32(n))
}

// Deterministically shuffle the result using the Fisher–Yates algorithm.
func shuffle(chars []byte, prng *SimplePRNG) {
	for i := len(chars) - 1; i > 0; i-- {
		j := prng.Intn(i + 1)
		chars[i], chars[j] = chars[j], chars[i]
	}
}

func hashFunction(input []byte) uint32 {
	hash := sha256.Sum256(input)
	seed := binary.BigEndian.Uint32(hash[:4])
	return seed
}

func Encode(input []byte, charSet *CharSet, limit int) ([]byte, error) {
	seed := hashFunction(input)
	prng := NewSimplePRNG(seed)

	union := ""
	// select one random character for each required group
	requiredChars := make([]byte, 0, len(charSet.Groups))
	for _, group := range charSet.Groups {
		union += group.Chars
		if group.Required {
			if len(group.Chars) == 0 {
				return nil, fmt.Errorf("required group %q has no characters", group.Name)
			}
			idx := prng.Intn(len(group.Chars))
			requiredChars = append(requiredChars, group.Chars[idx])
		}
	}

	if len(union) == 0 {
		return nil, fmt.Errorf("no characters available: union is empty")
	}

	if limit < len(requiredChars) {
		return nil, fmt.Errorf("password length %d is too short; must be at least %d", limit, len(requiredChars))
	}

	// Fill the remaining positions using the union
	remainingLength := limit - len(requiredChars)
	remainingChars := make([]byte, remainingLength)
	for i := 0; i < remainingLength; i++ {
		idx := prng.Intn(len(union))
		remainingChars[i] = union[idx]
	}

	passwordChars := append(requiredChars, remainingChars...)
	shuffle(passwordChars, prng)

	return passwordChars, nil
}
