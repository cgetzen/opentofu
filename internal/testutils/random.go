// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testutils

import (
	"hash/crc64"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"
)

// The functions below contain an assortment of random ID generation functions, partially ported and improved from the
// internal/legacy/helper/acctest package.

var randomSources = map[string]*rand.Rand{} //nolint:gochecknoglobals //This variable stores the randomness sources for DeterministicRandomID and needs to be global.
var randomLock = &sync.Mutex{}              //nolint:gochecknoglobals //This variable is required to lock the randomSources above.

// CharacterRange defines which characters to use for generating a random ID.
type CharacterRange string

const (
	CharacterRangeAlphaNumeric      CharacterRange = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	CharacterRangeAlphaNumericLower CharacterRange = "abcdefghijklmnopqrstuvwxyz0123456789"
	CharacterRangeAlphaNumericUpper CharacterRange = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	CharacterRangeAlpha             CharacterRange = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	CharacterRangeAlphaLower        CharacterRange = "abcdefghijklmnopqrstuvwxyz"
	CharacterRangeAlphaUpper        CharacterRange = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// DeterministicRandomID generates a pseudo-random identifier for the given test, using its name as a seed for
// randomness. This function guarantees that when queried in order, the values are always the same as long as the name
// of the test doesn't change.
func DeterministicRandomID(t *testing.T, length uint, characterSpace CharacterRange) string {
	var random *rand.Rand
	name := t.Name()
	var ok bool
	randomLock.Lock()
	random, ok = randomSources[name]
	if !ok {
		seed := crc64.Checksum([]byte(name), crc64.MakeTable(crc64.ECMA))
		random = rand.New(rand.NewSource(int64(seed))) //nolint:gosec //This random number generator is intentionally deterministic.
		randomSources[name] = random
		t.Cleanup(func() {
			randomLock.Lock()
			defer randomLock.Unlock()
			delete(randomSources, name)
		})
	}
	randomLock.Unlock()
	return RandomIDFromSource(random, length, characterSpace)
}

// RandomID returns a non-deterministic, pseudo-random identifier.
func RandomID(length uint, characterSpace CharacterRange) string {
	return RandomIDFromSource(rand.New(rand.NewSource(time.Now().UnixNano())), length, characterSpace) //nolint:gosec // Disabling gosec linting because this ID is for testing only.
}

// RandomIDPrefix returns a random identifier with a given prefix. The prefix length does not count towards the
// length.
func RandomIDPrefix(prefix string, length uint, characterSpace CharacterRange) string {
	return prefix + RandomID(length, characterSpace)
}

// RandomIDFromSource generates a random ID with the specified length based on the provided random parameter.
func RandomIDFromSource(random *rand.Rand, length uint, characterSpace CharacterRange) string {
	runes := []rune(characterSpace)
	var builder strings.Builder
	for i := uint(0); i < length; i++ {
		builder.WriteRune(runes[random.Intn(len(runes))])
	}
	return builder.String()
}
