package relayer

import (
	"math/big"

	"github.com/cosmos/cosmos-sdk/types"

	"github.com/ojo-network/cw-relayer/relayer/client"
	"github.com/ojo-network/cw-relayer/tools"
)

// genRateMsgs generates messages for exchange rates
func (r *Relayer) genRateMsgs(requestID uint64, resolveTime uint64) (msg []client.PriceFeedData) {
	for _, rate := range r.exchangeRates {
		var byteArray [32]byte
		copy(byteArray[:], rate.Denom)
		msg = append(msg, client.PriceFeedData{
			AssetName:   byteArray,
			Value:       DecTofactorBigInt(rate.Amount),
			Id:          big.NewInt(int64(requestID)),
			ResolveTime: big.NewInt(int64(resolveTime)),
		})
	}
	return
}

// genDeviationsMsg generates messages for deviations
func (r *Relayer) genDeviationsMsg(requestID uint64, resolveTime uint64) (msg []client.PriceFeedData) {
	for _, rate := range r.historicalDeviations {
		byteArray := tools.StringToByte32(rate.Denom)
		msg = append(msg, client.PriceFeedData{
			AssetName:   byteArray,
			Value:       DecTofactorBigInt(rate.Amount),
			Id:          big.NewInt(int64(requestID)),
			ResolveTime: big.NewInt(int64(resolveTime)),
		})
	}

	return
}

// genMedianMsg generates messages for medians by collecting values of denoms
func (r *Relayer) genMedianMsg(requestID uint64, resolveTime uint64) (msg []client.PriceFeedMedianData) {
	medianRates := map[[32]byte][]*big.Int{}
	for _, rate := range r.historicalMedians {
		byteArray := tools.StringToByte32(rate.Denom)
		medianRates[byteArray] = append(medianRates[byteArray], DecTofactorBigInt(rate.Amount))
	}

	for symbol, rates := range medianRates {
		msg = append(msg, client.PriceFeedMedianData{
			AssetName:   symbol,
			Values:      rates,
			ResolveTime: big.NewInt(int64(resolveTime)),
			Id:          big.NewInt(int64(requestID)),
		})
	}

	return
}

// DecTofactorBigInt multiplies amount by rate factor to make it compatible with contracts
func DecTofactorBigInt(amount types.Dec) *big.Int {
	return amount.Mul(RateFactor).TruncateInt().BigInt()
}
