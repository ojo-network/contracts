package server

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"strconv"

	"github.com/cosmos/cosmos-sdk/types"
	oracletypes "github.com/ojo-network/ojo/x/oracle/types"
	"google.golang.org/grpc"
)

type Server struct {
	oracletypes.UnimplementedQueryServer
	mockPrices  []types.DecCoin
	priceStamps []oracletypes.PriceStamp
	priceMap    map[string][]oracletypes.PriceStamp
}

func (s *Server) ExchangeRates(context.Context, *oracletypes.QueryExchangeRates) (*oracletypes.QueryExchangeRatesResponse, error) {
	return &oracletypes.QueryExchangeRatesResponse{
		ExchangeRates: s.mockPrices,
	}, nil
}

func (s *Server) Medians(context.Context, *oracletypes.QueryMedians) (*oracletypes.QueryMediansResponse, error) {
	return &oracletypes.QueryMediansResponse{
		Medians: s.priceStamps,
	}, nil
}

func (s *Server) MedianDeviations(context.Context, *oracletypes.QueryMedianDeviations) (*oracletypes.QueryMedianDeviationsResponse, error) {
	return &oracletypes.QueryMedianDeviationsResponse{
		MedianDeviations: s.priceStamps,
	}, nil
}

func (s *Server) setMockPrices() {
	s.priceMap = make(map[string][]oracletypes.PriceStamp)
	for i := 0; i < 5; i++ {
		price := rand.Float64()
		if i%2 == 1 {
			// to have prices above 1
			price = price * 100000
		}

		denom := fmt.Sprintf("TEST-%v", i)
		priceDec := types.MustNewDecFromStr(strconv.FormatFloat(price, 'f', 9, 64))
		exchangeRate := types.NewDecCoinFromDec(denom, priceDec)
		s.mockPrices = append(s.mockPrices, exchangeRate)

		priceStamps := make([]oracletypes.PriceStamp, 10)
		for j := 0; j < 10; j++ {
			priceStamps[j] = oracletypes.PriceStamp{
				ExchangeRate: &exchangeRate,
				BlockNum:     uint64(j),
			}
		}

		s.priceMap[denom] = priceStamps
		s.priceStamps = append(s.priceStamps, priceStamps...)
	}
}

func (s *Server) InitMockPriceServer(grpcPort string) error {
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		return err
	}

	s.setMockPrices()
	grpcServer := grpc.NewServer()
	oracletypes.RegisterQueryServer(grpcServer, s)

	go func() {
		err := grpcServer.Serve(lis)
		if err != nil {
			return
		}
	}()

	return nil
}

func (s *Server) GetMockPrices() []types.DecCoin {
	return s.mockPrices
}

func (s *Server) GetPriceStamps(denom string) ([]oracletypes.PriceStamp, error) {
	data, found := s.priceMap[denom]
	if !found {
		return nil, fmt.Errorf("denom not found")
	}

	return data, nil
}
