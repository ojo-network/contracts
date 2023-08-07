package e2e

import (
	"context"
	"encoding/json"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ojo-network/cw-relayer/relayer"
	"github.com/ojo-network/cw-relayer/tests/e2e/orchestrator"
)

type (
	LastPing struct {
		Relayer Relayer `json:"last_ping"`
	}

	Relayer struct {
		Relayer string `json:"relayer"`
	}
)

func (s *IntegrationTestSuite) TestPriceCallback() {
	grpcConn, err := grpc.Dial(
		s.orchestrator.QueryRpc,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	s.Require().NoError(err)
	defer grpcConn.Close()
	queryClient := wasmtypes.NewQueryClient(grpcConn)
	for _, rate := range s.priceServer.GetMockPrices() {
		rate := rate
		err := s.orchestrator.RequestMsg(orchestrator.Price, rate.Denom)
		s.Require().NoError(err)

		s.Require().Eventually(
			func() bool {
				queryID := s.orchestrator.GenerateRequestIDQuery(orchestrator.Price, rate.Denom)
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				resp, err := queryClient.SmartContractState(ctx, queryID)
				if err != nil || len(resp.String()) == 0 {
					return false
				}

				query := s.orchestrator.GeneratePriceQuery(orchestrator.Price, rate.Denom)
				ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				resp, err = queryClient.SmartContractState(ctx, query)
				if err != nil || len(resp.String()) == 0 {
					return false
				}

				var callbackRate string
				err = json.Unmarshal(resp.Data, &callbackRate)
				if err != nil {
					return false
				}

				if callbackRate == rate.Amount.Mul(relayer.RateFactor).TruncateInt().String() {
					return true
				}

				return false
			},
			7*time.Minute, 20*time.Second,
			"rate request and callback failed")
	}
}

func (s *IntegrationTestSuite) TestMedianCallback() {
	grpcConn, err := grpc.Dial(
		s.orchestrator.QueryRpc,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	s.Require().NoError(err)
	defer grpcConn.Close()
	queryClient := wasmtypes.NewQueryClient(grpcConn)

	for _, rate := range s.priceServer.GetMockPrices() {
		rate := rate
		err = s.orchestrator.RequestMsg(orchestrator.Median, rate.Denom)
		s.Require().NoError(err)

		s.Require().Eventually(
			func() bool {
				queryID := s.orchestrator.GenerateRequestIDQuery(orchestrator.Median, rate.Denom)
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				resp, err := queryClient.SmartContractState(ctx, queryID)
				if err != nil || len(resp.String()) == 0 {
					return false
				}

				query := s.orchestrator.GeneratePriceQuery(orchestrator.Median, rate.Denom)
				ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				resp, err = queryClient.SmartContractState(ctx, query)
				if err != nil || len(resp.String()) == 0 {
					return false
				}

				var callbackRates []string
				err = json.Unmarshal(resp.Data, &callbackRates)
				if err != nil {
					return false
				}

				if len(callbackRates) != 0 {
					priceStamps, err := s.priceServer.GetPriceStamps(rate.Denom)
					if err != nil {
						return false
					}

					if len(priceStamps) != len(callbackRates) {
						return false
					}

					for i, priceStamp := range priceStamps {
						if callbackRates[i] != priceStamp.ExchangeRate.Amount.Mul(relayer.RateFactor).TruncateInt().String() {
							return false
						}
					}

					return true
				}

				return false

			},
			7*time.Minute, 20*time.Second,
			"median request and callback failed")
	}
}

func (s *IntegrationTestSuite) TestDeviationCallback() {
	grpcConn, err := grpc.Dial(
		s.orchestrator.QueryRpc,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	s.Require().NoError(err)
	defer grpcConn.Close()
	queryClient := wasmtypes.NewQueryClient(grpcConn)

	for _, rate := range s.priceServer.GetMockPrices() {
		rate := rate
		err = s.orchestrator.RequestMsg(orchestrator.Deviation, rate.Denom)
		s.Require().NoError(err)

		s.Require().Eventually(
			func() bool {
				queryID := s.orchestrator.GenerateRequestIDQuery(orchestrator.Deviation, rate.Denom)
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				resp, err := queryClient.SmartContractState(ctx, queryID)
				if err != nil || len(resp.String()) == 0 {
					return false
				}

				query := s.orchestrator.GeneratePriceQuery(orchestrator.Deviation, rate.Denom)
				ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				resp, err = queryClient.SmartContractState(ctx, query)
				if err != nil || len(resp.String()) == 0 {
					return false
				}

				var callbackRates []string
				err = json.Unmarshal(resp.Data, &callbackRates)
				if err != nil {
					return false
				}

				if len(callbackRates) != 0 {
					priceStamps, err := s.priceServer.GetPriceStamps(rate.Denom)
					if err != nil {
						return false
					}

					if len(priceStamps) != len(callbackRates) {
						return false
					}

					for i, priceStamp := range priceStamps {
						if callbackRates[i] != priceStamp.ExchangeRate.Amount.Mul(relayer.RateFactor).TruncateInt().String() {
							return false
						}
					}

					return true
				}

				return false
			},
			7*time.Minute, 20*time.Second,
			"deviation request and callback failed")
	}
}
