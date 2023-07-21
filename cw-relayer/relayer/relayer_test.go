package relayer

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"

	"github.com/ojo-network/cw-relayer/relayer/client"
)

type RelayerTestSuite struct {
	suite.Suite
	relayer         *Relayer
	msg             chan types.Msg
	priceService    PriceService
	denomList       []string
	clientRequest   map[string][]client.PriceRequest
	timestamp       string
	contractAddress string
}

// SetupClientRequest setup client requests for rate, median and deviation, for given denom list
func (rts *RelayerTestSuite) SetupClientRequest() {
	rts.clientRequest = make(map[string][]client.PriceRequest)
	request := client.PriceRequest{
		EventContractAddress: rts.contractAddress,
		ResolveTime:          rts.timestamp,
		RequestedSymbol:      "",
		CallbackData:         "sig",
		CallbackSig:          "callback",
		RequestID:            "rate-",
	}

	for i, denom := range rts.denomList {
		request.Event = client.RequestRate
		request.RequestedSymbol = denom
		request.RequestID = "rate-" + strconv.Itoa(i)
		rts.clientRequest[denom] = append(rts.clientRequest[denom], request)

		request.Event = client.RequestMedian
		request.RequestID = "median-" + strconv.Itoa(i)
		rts.clientRequest[denom] = append(rts.clientRequest[denom], request)

		request.Event = client.RequestDeviation
		request.RequestID = "deviation-" + strconv.Itoa(i)
		rts.clientRequest[denom] = append(rts.clientRequest[denom], request)

	}
	fmt.Println(rts.clientRequest)
}

func (rts *RelayerTestSuite) SetupMockPriceService() *PriceService {
	mockService := &PriceService{}
	mockService.exchangeRates = make(map[string]Price)
	mockService.medianRates = make(map[string]Median)
	mockService.deviationRates = make(map[string]Deviation)

	min := 1
	// rate factor
	max := 1000000000
	for _, denom := range rts.denomList {
		price := rand.Intn(max-min+1) + min
		deviation := rand.Intn(max-min+1) + min
		median := rand.Intn(max-min+1) + min

		mockService.exchangeRates[denom] = Price{
			Price:     strconv.Itoa(price),
			Timestamp: rts.timestamp,
		}

		mockService.deviationRates[denom] = Deviation{
			Deviation: strconv.Itoa(deviation),
			Timestamp: rts.timestamp,
		}

		mockService.medianRates[denom] = Median{
			Median:    []string{strconv.Itoa(median)},
			Timestamp: rts.timestamp,
		}
	}

	return mockService
}

func (rts *RelayerTestSuite) SetupSuite() {
	rts.timestamp = strconv.Itoa(time.Now().Second())
	rts.denomList = []string{"ATOM", "OJO", "UMEE", "TEST"}
	rts.SetupClientRequest()
	mockService := rts.SetupMockPriceService()
	rts.msg = make(chan types.Msg, 10000)
	rts.relayer = New(
		zerolog.Nop(),
		&client.ContractSubscribe{},
		mockService,
		"",
		"",
		0,
		1*time.Second,
		nil,
		rts.msg,
	)
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(RelayerTestSuite))
}

func (rts *RelayerTestSuite) TestStop() {
	rts.Eventually(
		func() bool {
			rts.relayer.Stop()
			return true
		},
		5*time.Second,
		time.Second,
	)
}

func (rts *RelayerTestSuite) Test_processRequests() {
	total := 0
	rts.Eventually(func() bool {
		for msg := range rts.msg {
			msg := msg
			wasmMsg, _ := msg.(*wasmtypes.MsgExecuteContract)
			var jsonMsg map[string]interface{}
			err := json.Unmarshal(wasmMsg.Msg, &jsonMsg)
			rts.Require().NoError(err)

			fmt.Println(jsonMsg["callback"])

			total += 1
			if total >= 10 {
				return true
			}
		}

		return true
	},
		30*time.Second,
		2*time.Second,
	)

	err := rts.relayer.processRequests(context.Background(), rts.clientRequest)
	rts.Require().NoError(err)
}
