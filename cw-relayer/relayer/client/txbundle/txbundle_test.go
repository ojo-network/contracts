package txbundle

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	"github.com/ojo-network/cw-relayer/relayer"
)

func generateMockMessages(n int) []sdk.Msg {
	msgs := make([]sdk.Msg, n)
	for i := 0; i < n; i++ {
		msgData, _ := json.Marshal(relayer.CallbackRate{Req: relayer.CallbackData{
			RequestID:    strconv.Itoa(i),
			Symbol:       fmt.Sprintf("TEST-%v", i),
			SymbolRate:   strconv.Itoa(i * 100),
			LastUpdated:  strconv.Itoa(time.Now().Second()),
			CallbackData: []byte("testcallback"),
		}})

		msgs[i] = &wasmtypes.MsgExecuteContract{
			Sender:   "sender",
			Contract: "contact",
			Msg:      msgData,
			Funds:    nil,
		}
	}

	return msgs
}

func TestTxbundle_Bundler_Without_Estimate(t *testing.T) {
	tx := Txbundle{}
	tx.timeoutDuration = 1 * time.Minute
	tx.estimateAndBundle = false
	tx.totalTxThreshold = 100
	tx.MsgChan = make(chan sdk.Msg, 1000)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error {
		return tx.Bundler(ctx)
	})

	msgs := generateMockMessages(10)
	for _, msg := range msgs {
		tx.MsgChan <- msg
	}

	require.NoError(t, g.Wait())

	// check msg
	require.Len(t, tx.msgs, len(msgs))
}
