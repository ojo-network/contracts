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

const QUERY_PORT = "9090"

type Server struct {
	oracletypes.UnimplementedQueryServer
	mockPrices  []types.DecCoin
	priceStamps []oracletypes.PriceStamp
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
	for i := 0; i < 5; i++ {
		price := rand.Float64()
		if i%2 == 1 {
			// to have prices above 1
			price = price * 100000
		}

		priceDec := types.MustNewDecFromStr(strconv.FormatFloat(price, 'f', 9, 64))
		exchangeRate := types.NewDecCoinFromDec(fmt.Sprintf("TEST-%v", i), priceDec)
		s.mockPrices = append(s.mockPrices, exchangeRate)
		s.priceStamps = append(s.priceStamps, oracletypes.PriceStamp{
			ExchangeRate: &exchangeRate,
			BlockNum:     uint64(i),
		})
	}
}

func (s *Server) InitMockPriceServer() error {
	lis, err := net.Listen("tcp", ":"+QUERY_PORT)
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
