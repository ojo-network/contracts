package relayer

import (
	"math/big"

	"github.com/ojo-network/cw-relayer/relayer/client"
	"github.com/ojo-network/cw-relayer/tools"
)

type MsgType int

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
		byteArray := tools.StringToByte32(rate.Denom)
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
		byteArray := tools.StringToByte32(rate.Denom)
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
