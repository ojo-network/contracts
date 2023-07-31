package relayer

import (
	"context"
	"encoding/json"
	"math/rand"
	"strconv"
	"sync"
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
	mut             sync.Mutex
	relayer         *Relayer
	msg             chan types.Msg
	priceService    *PriceService
	denomList       []string
	clientRequest   map[string][]client.PriceRequest
	requestMap      map[string]map[string]client.PriceRequest
	timestamp       string
	contractAddress string
}

// SetupClientRequest setup client requests for rate, median and deviation, for given denom list
func (rts *RelayerTestSuite) SetupClientRequest() {
	rts.mut.Lock()
	defer rts.mut.Unlock()

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
		rts.requestMap[denom][request.RequestID] = request

		request.Event = client.RequestMedian
		request.RequestID = "median-" + strconv.Itoa(i)
		rts.clientRequest[denom] = append(rts.clientRequest[denom], request)
		rts.requestMap[denom][request.RequestID] = request

		request.Event = client.RequestDeviation
		request.RequestID = "deviation-" + strconv.Itoa(i)
		rts.clientRequest[denom] = append(rts.clientRequest[denom], request)
		rts.requestMap[denom][request.RequestID] = request
	}
}

func (rts *RelayerTestSuite) SetupMockPriceService() *PriceService {
	rts.mut.Lock()
	defer rts.mut.Unlock()
	mockService := &PriceService{}
	mockService.exchangeRates = make(map[string]Price)
	mockService.medianRates = make(map[string]Median)
	mockService.deviationRates = make(map[string]Deviation)

	min := 1
	// rate factor
	max := 1000000000
	for _, denom := range rts.denomList {
		// init request map
		rts.requestMap[denom] = make(map[string]client.PriceRequest)

		price := rand.Intn(max-min+1) + min
		median := rand.Intn(max-min+1) + min

		mockService.exchangeRates[denom] = Price{
			Price:     strconv.Itoa(price),
			Timestamp: rts.timestamp,
		}

		mockService.deviationRates[denom] = Deviation{
			Deviation: []string{strconv.Itoa(price)},
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
	rts.requestMap = make(map[string]map[string]client.PriceRequest)

	mockService := rts.SetupMockPriceService()
	rts.SetupClientRequest()
	rts.priceService = mockService
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

	close(rts.msg)
}

func (rts *RelayerTestSuite) Test_processRequests() {
	total := 0
	go rts.relayer.processRequests(context.Background(), rts.clientRequest) //nolint
	for msg := range rts.msg {
		msg := msg
		wasmMsg, _ := msg.(*wasmtypes.MsgExecuteContract)
		rts.Require().Equal(wasmMsg.Contract, rts.contractAddress)

		var jsonMsg map[string]map[string]interface{}
		err := json.Unmarshal(wasmMsg.Msg, &jsonMsg)
		rts.Require().NoError(err)

		callback, err := parseAndCheck(jsonMsg["callback"])
		rts.Require().NoError(err)

		switch callback := callback.(type) {
		case CallbackData:
			data := callback
			request := rts.requestMap[data.Symbol][data.RequestID]
			rate := rts.priceService.exchangeRates[data.Symbol]
			rts.Require().Equal(data.RequestID, request.RequestID)
			rts.Require().Equal(string(data.CallbackData), request.CallbackData)
			rts.Require().Equal(data.SymbolRate, rate.Price)
			rts.Require().Equal(data.LastUpdated, rate.Timestamp)
			total += 1

		case CallbackDataHistorical:
			data := callback
			request := rts.requestMap[data.Symbol][data.RequestID]
			rate := rts.priceService.medianRates[data.Symbol]
			rts.Require().Equal(data.RequestID, request.RequestID)
			rts.Require().Equal(string(data.CallbackData), request.CallbackData)
			rts.Require().Equal(data.SymbolRates, rate.Median)
			rts.Require().Equal(data.LastUpdated, rate.Timestamp)
			total += 1
		}

		if total == len(rts.clientRequest) {
			return
		}
	}
}

func parseAndCheck(msg map[string]interface{}) (interface{}, error) {
	keys := []string{"callback_rate_deviation", "callback_rate_data", "callback_rate_median"}
	for i, key := range keys {
		msg, found := msg[key]
		if !found {
			continue
		}

		data, err := json.Marshal(msg)
		if err != nil {
			return nil, err
		}

		if i <= 1 {
			var callback CallbackData
			err = json.Unmarshal(data, &callback)
			if err != nil {
				return nil, err
			}

			return callback, nil
		} else {
			var callback CallbackDataHistorical
			err = json.Unmarshal(data, &callback)
			if err != nil {
				return nil, err
			}

			return callback, nil
		}
	}

	return nil, nil
}
