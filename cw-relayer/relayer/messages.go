package relayer

import (
	"math/big"

	"github.com/ojo-network/cw-relayer/relayer/client"
)

type MsgType int

const (
	RelayRate MsgType = iota + 1
	RelayHistoricalMedian
	RelayHistoricalDeviation
	QueryRateMsg
	QueryMedianRateMsg
	QueryDeviationRateMsg
)

func (m MsgType) String() string {
	return [...]string{"relay", "relay_historical_median", "relay_historical_deviation"}[m-1]
}

func (r *Relayer) genRateMsgs(requestID uint64, resolveTime uint64) (msg []client.PriceFeedData) {
	for _, rate := range r.exchangeRates {
		var byteArray [32]byte
		copy(byteArray[:], rate.Denom)
		msg = append(msg, client.PriceFeedData{
			Name:        byteArray,
			Value:       rate.Amount.Mul(RateFactor).TruncateInt().BigInt(),
			Id:          big.NewInt(int64(requestID)),
			ResolveTime: big.NewInt(int64(resolveTime)),
		})
	}
	return
}

func (r *Relayer) genDeviationsMsg(requestID uint64, resolveTime uint64) (msg []client.PriceFeedData) {
	for _, rate := range r.historicalDeviations {
		var byteArray [32]byte
		copy(byteArray[:], rate.Denom)
		msg = append(msg, client.PriceFeedData{
			Name:        byteArray,
			Value:       rate.Amount.Mul(RateFactor).TruncateInt().BigInt(),
			Id:          big.NewInt(int64(requestID)),
			ResolveTime: big.NewInt(int64(resolveTime)),
		})
	}

	return
}

func (r *Relayer) genMedianMsg(requestID uint64, resolveTime uint64) (msg []client.PriceFeedMedianData) {
	medianRates := map[[32]byte][]*big.Int{}
	for _, rate := range r.historicalMedians {
		var byteArray [32]byte
		copy(byteArray[:], rate.Denom)
		medianRates[byteArray] = append(medianRates[byteArray], rate.Amount.Mul(RateFactor).TruncateInt().BigInt())
	}

	for symbol, rates := range medianRates {
		msg = append(msg, client.PriceFeedMedianData{
			Name:        symbol,
			Value:       rates,
			ResolveTime: big.NewInt(int64(requestID)),
			Id:          big.NewInt(int64(resolveTime)),
		})
	}

	return
}
