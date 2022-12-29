package orchestrator

import (
	"context"
	"net"

	"github.com/cosmos/cosmos-sdk/types"
	oracletypes "github.com/ojo-network/ojo/x/oracle/types"
	"google.golang.org/grpc"
)

type server struct {
	oracletypes.UnimplementedQueryServer
}

func (s *server) ExchangeRates(context.Context, *oracletypes.QueryExchangeRates) (*oracletypes.QueryExchangeRatesResponse, error) {
	mockPrices := []types.DecCoin{
		types.NewDecCoinFromDec("TEST", types.MustNewDecFromStr("1.23456789")),
		types.NewDecCoinFromDec("STAKE", types.MustNewDecFromStr("0.00123456")),
	}
	return &oracletypes.QueryExchangeRatesResponse{
		ExchangeRates: types.NewDecCoins(mockPrices...),
	}, nil
}

func (o *Orchestrator) InitMockPriceServer(grpcPort string) error {
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	oracletypes.RegisterQueryServer(s, &server{})

	go func() {
		err := s.Serve(lis)
		if err != nil {
			return
		}
	}()

	return nil
}
