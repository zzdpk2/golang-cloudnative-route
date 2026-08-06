// Package prefixcache models the prompt-prefix keys used to reason about KV
// cache affinity. It deliberately stops before scheduling: the output is an
// immutable request property consumed by later Router policy.
package prefixcache

// DefaultBlockSize is the learning fixture's KV-cache block size.
const DefaultBlockSize = 16

// Token stands in for a tokenizer output ID. The learning tokenizer is byte
// sensitive so that UTF-8 input exposes the same boundary problem a Router
// must handle when its tokenizer differs from a Model Server tokenizer.
type Token int32

// BlockHash identifies one complete block in its exact prompt-prefix context.
type BlockHash uint64

// Tokenize returns a deterministic token sequence for the input bytes.
//
// Hint: confirm what a Go range loop iterates before choosing it here.
func Tokenize(text string) []Token {
	panic("TODO")
}

// Chunk returns consecutive complete blocks and excludes a trailing partial
// block. A non-positive block size is invalid.
//
// Hint: decide and document whether callers may observe aliasing through a
// shared backing array.
func Chunk(tokens []Token, blockSize int) [][]Token {
	panic("TODO")
}

// HashBlocks returns one deterministic key per block. A key must change when
// either the current block or any earlier block changes, while a shared prompt
// prefix must retain the same leading keys.
//
// Hint: test the failure mode where two prompts contain the same block after
// different prefixes; a reusable KV entry is valid only from the beginning.
func HashBlocks(blocks [][]Token) []BlockHash {
	panic("TODO")
}

// BlocksForPrompt applies the package contracts as one prompt-to-keys pipeline.
// Invalid block sizes follow Chunk's invalid-input result.
func BlocksForPrompt(text string, blockSize int) []BlockHash {
	panic("TODO")
}

// MatchedPrefix reports the number of equal leading keys and stops at the
// first difference.
func MatchedPrefix(a, b []BlockHash) int {
	panic("TODO")
}
