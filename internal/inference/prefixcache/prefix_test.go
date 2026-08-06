package prefixcache

import (
	"reflect"
	"testing"
)

func TestPrefixCacheFoundations(t *testing.T) {
	t.Run("tokenization follows bytes", func(t *testing.T) {
		if got, want := Tokenize("Aé"), []Token{65, 195, 169}; !reflect.DeepEqual(got, want) {
			t.Fatalf("Tokenize() = %v, want %v", got, want)
		}
		if got := Tokenize(""); len(got) != 0 {
			t.Fatalf("Tokenize(empty) = %v, want empty", got)
		}
	})

	t.Run("only complete blocks are cacheable", func(t *testing.T) {
		tokens := []Token{1, 2, 3, 4, 5}
		got := Chunk(tokens, 2)
		want := [][]Token{{1, 2}, {3, 4}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("Chunk() = %v, want %v", got, want)
		}
		if got := Chunk(tokens, 0); got != nil {
			t.Fatalf("Chunk() with invalid size = %v, want nil", got)
		}
		if got := Chunk(tokens, -1); got != nil {
			t.Fatalf("Chunk() with negative size = %v, want nil", got)
		}
		tokens[0] = 99
		if got[0][0] != 1 {
			t.Fatalf("Chunk() leaked the caller's backing array: %v", got)
		}
		got[1][0] = 88
		if tokens[2] != 3 {
			t.Fatal("mutating a returned block changed the input tokens")
		}
	})

	t.Run("chained hashes are deterministic and prefix-sensitive", func(t *testing.T) {
		left := [][]Token{{1, 2}, {3, 4}}
		samePrefix := [][]Token{{1, 2}, {3, 4}, {5, 6}}
		differentPrefix := [][]Token{{9, 9}, {3, 4}}

		leftHashes := HashBlocks(left)
		if again := HashBlocks(left); !reflect.DeepEqual(leftHashes, again) {
			t.Fatalf("same blocks produced %v then %v", leftHashes, again)
		}
		longerHashes := HashBlocks(samePrefix)
		if len(leftHashes) != 2 || !reflect.DeepEqual(leftHashes, longerHashes[:2]) {
			t.Fatalf("shared prefix hashes = %v and %v", leftHashes, longerHashes)
		}
		otherHashes := HashBlocks(differentPrefix)
		if leftHashes[1] == otherHashes[1] {
			t.Fatal("the same trailing block under a different prefix must hash differently")
		}
		differentCurrentBlock := HashBlocks([][]Token{{1, 2}, {8, 8}})
		if leftHashes[1] == differentCurrentBlock[1] {
			t.Fatal("a changed current block under the same prefix must change its hash")
		}
		if got := HashBlocks(nil); len(got) != 0 {
			t.Fatalf("HashBlocks(nil) = %v, want empty", got)
		}
	})

	t.Run("the prompt pipeline drops its partial tail", func(t *testing.T) {
		text := "0123456789"
		got := BlocksForPrompt(text, 4)
		if len(got) != 2 {
			t.Fatalf("BlocksForPrompt() returned %d hashes, want 2", len(got))
		}
		if want := HashBlocks(Chunk(Tokenize(text), 4)); !reflect.DeepEqual(got, want) {
			t.Fatalf("BlocksForPrompt() = %v, want %v", got, want)
		}
		if got := BlocksForPrompt(text, 0); got != nil {
			t.Fatalf("BlocksForPrompt() with invalid size = %v, want nil", got)
		}
	})

	t.Run("matching stops at the first different block", func(t *testing.T) {
		cases := []struct {
			name  string
			left  []BlockHash
			right []BlockHash
			want  int
		}{
			{name: "same", left: []BlockHash{1, 2, 3}, right: []BlockHash{1, 2, 3}, want: 3},
			{name: "shorter", left: []BlockHash{1, 2}, right: []BlockHash{1, 2, 3}, want: 2},
			{name: "different tail", left: []BlockHash{1, 2, 3}, right: []BlockHash{1, 2, 9}, want: 2},
			{name: "shared tail only", left: []BlockHash{1, 2, 3}, right: []BlockHash{9, 2, 3}, want: 0},
			{name: "both empty", left: nil, right: nil, want: 0},
			{name: "one empty", left: nil, right: []BlockHash{1}, want: 0},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				if got := MatchedPrefix(tc.left, tc.right); got != tc.want {
					t.Fatalf("MatchedPrefix() = %d, want %d", got, tc.want)
				}
			})
		}
	})
}
