package blockcache_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gacevicljubisa/swaplist/pkg/blockcache"
)

func TestBasic(t *testing.T) {
	t.Parallel()
	c := blockcache.New()
	b := &types.Block{}
	c.Set(b)
	if c.Get(b.Hash()).Hash() != b.Hash() {
		t.Fatal("expected block to be cached")
	}
}
