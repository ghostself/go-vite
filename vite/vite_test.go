package vite

import (
	"flag"
	"fmt"
	"testing"

	"math/big"

	"github.com/vitelabs/go-vite/chain"
	"github.com/vitelabs/go-vite/common"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/config"
	"github.com/vitelabs/go-vite/crypto/ed25519"
	"github.com/vitelabs/go-vite/generator"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/pool"
	"github.com/vitelabs/go-vite/pow"
	"github.com/vitelabs/go-vite/verifier"
	"github.com/vitelabs/go-vite/vm/contracts"
	"github.com/vitelabs/go-vite/wallet"
)

var genesisAccountPrivKeyStr string

func init() {
	flag.StringVar(&genesisAccountPrivKeyStr, "k", "", "")
	flag.Parse()
	fmt.Println(genesisAccountPrivKeyStr)

}

func PrepareVite() (chain.Chain, *generator.Generator, *verifier.AccountVerifier, pool.BlockPool) {
	c := chain.NewChain(&config.Config{DataDir: common.DefaultDataDir()})
	c.Init()
	c.Start()

	w := wallet.New(nil)
	g := generator.NewGenerator(c, w.KeystoreManager)

	v := verifier.NewAccountVerifier(c, nil)

	p := pool.NewPool(c)

	p.Init(&pool.MockSyncer{}, w, verifier.NewSnapshotVerifier(c, nil), v)
	p.Start()

	return c, g, v, p
}

func TestSend(t *testing.T) {
	c, g, v, p := PrepareVite()
	genesisAccountPrivKey, _ := ed25519.HexToPrivateKey(genesisAccountPrivKeyStr)
	genesisAccountPubKey := genesisAccountPrivKey.PubByte()
	fromBlock, err := c.GetLatestAccountBlock(&contracts.AddressMintage)
	if err != nil {
		fmt.Println(err)
		return
	}
	block := &ledger.AccountBlock{
		Height:         1,
		AccountAddress: ledger.GenesisAccountAddress,
		FromBlockHash:  fromBlock.Hash,
		BlockType:      ledger.BlockTypeReceive,
		Fee:            big.NewInt(0),
		Amount:         big.NewInt(0),
		TokenId:        ledger.ViteTokenId,
		SnapshotHash:   c.GetLatestSnapshotBlock().Hash,
		Timestamp:      c.GetLatestSnapshotBlock().Timestamp,
		PublicKey:      genesisAccountPubKey,
	}

	nonce := pow.GetPowNonce(nil, types.DataHash(append(block.AccountAddress.Bytes(), block.PrevHash.Bytes()...)))
	block.Nonce = nonce[:]
	block.Hash = block.ComputeHash()
	block.Signature = ed25519.Sign(genesisAccountPrivKey, block.Hash.Bytes())

	blocks, err := v.VerifyforRPC(block, g)
	t.Log(blocks, err)

	t.Log(blocks[0].VmContext.GetBalance(&ledger.GenesisAccountAddress, &ledger.ViteTokenId), err)

	t.Log(blocks[0].AccountBlock.Hash)

	p.AddDirectAccountBlock(ledger.GenesisAccountAddress, blocks[0])

	accountBlock, e := c.GetLatestAccountBlock(&ledger.GenesisAccountAddress)

	t.Log(accountBlock.Hash, e)
}