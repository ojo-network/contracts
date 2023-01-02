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
	mockPrices []types.DecCoin
}

func (s *Server) ExchangeRates(context.Context, *oracletypes.QueryExchangeRates) (*oracletypes.QueryExchangeRatesResponse, error) {
	return &oracletypes.QueryExchangeRatesResponse{
		ExchangeRates: s.mockPrices,
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
		s.mockPrices = append(s.mockPrices, types.NewDecCoinFromDec(fmt.Sprintf("TEST-%v", i), priceDec))
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
