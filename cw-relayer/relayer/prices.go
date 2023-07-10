package relayer

import (
	"context"
	"fmt"
	"github.com/ojo-network/cw-relayer/tools"
	oracletypes "github.com/ojo-network/ojo/x/oracle/types"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"strconv"
	"sync"
	"time"
)

type PriceService struct {
	logger          zerolog.Logger
	queryRPCS       []string
	maxQueryRetries int64
	queryRetries    int64
	index           int

	exchangeRates  map[string]Price
	deviationRates map[string]Deviation
	medianRates    map[string]Median

	queryTimeout time.Duration

	mut sync.Mutex

	eventTick chan struct{}
}

type Price struct {
	Price     string
	Timestamp string
}

type Deviation struct {
	Deviation string
	Timestamp string
}

type Median struct {
	Median    []string
	Timestamp string
}

func NewPriceService(
	logger zerolog.Logger,
	queryRPCS []string,
	maxQueryRetries int64,
	queryTimeout time.Duration,
	eventTick chan struct{},
) *PriceService {
	return &PriceService{
		logger:          logger.With().Str("module", "price service").Logger(),
		queryRPCS:       queryRPCS,
		maxQueryRetries: maxQueryRetries,
		eventTick:       eventTick,
		mut:             sync.Mutex{},
		queryTimeout:    queryTimeout,
	}
}

func (p *PriceService) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-p.eventTick:
			p.logger.Info().Msg("oracle price tick")
			err := p.setDenomPrices(ctx)
			if err != nil {
				p.logger.Error().Err(err).Msg("error configuring prices")
			}
		}
	}
}

func (p *PriceService) setDenomPrices(ctx context.Context) error {
	if p.queryRetries > p.maxQueryRetries {
		p.queryRetries = 0
		return fmt.Errorf("retry threshold exceeded")
	}

	grpcConn, err := grpc.Dial(
		p.queryRPCS[p.index],
		// the Cosmos SDK doesn't support any transport security mechanism
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(tools.DialerFunc),
	)

	// retry or switch rpc
	if err != nil {
		p.increment()
		return p.setDenomPrices(ctx)
	}

	defer grpcConn.Close()

	queryClient := oracletypes.NewQueryClient(grpcConn)

	ctx, cancel := context.WithTimeout(ctx, p.queryTimeout)
	defer cancel()

	queryResponse, err := queryClient.ExchangeRates(ctx, &oracletypes.QueryExchangeRates{})
	// assuming an issue with rpc if exchange rates are empty
	if err != nil || queryResponse.ExchangeRates.Empty() {
		p.logger.Debug().Msg("error querying exchange rates")
		p.increment()
		return p.setDenomPrices(ctx)
	}

	for _, coin := range queryResponse.ExchangeRates {
		p.exchangeRates[coin.Denom] = Price{
			Price:     coin.Amount.Mul(RateFactor).TruncateInt().String(),
			Timestamp: strconv.Itoa(int(time.Now().Unix())),
		}
	}

	g, _ := errgroup.WithContext(ctx)
	g.Go(
		func() error {
			deviationsQueryResponse, err := queryClient.MedianDeviations(ctx, &oracletypes.QueryMedianDeviations{})
			if err != nil {
				return err
			}

			if len(deviationsQueryResponse.MedianDeviations) == 0 {
				return noDeviations
			}

			deviations := make(map[string]Deviation)
			for _, priceStamp := range deviationsQueryResponse.MedianDeviations {
				denom := priceStamp.ExchangeRate.Denom
				amount := priceStamp.ExchangeRate.Amount.Mul(RateFactor).TruncateInt()
				if _, found := deviations[denom]; !found {
					deviations[denom] = Deviation{
						Deviation: amount.String(),
						Timestamp: strconv.Itoa(int(time.Now().Unix())),
					}
				} else {
					p.logger.Err(fmt.Errorf("multiple deviations found"))
				}
			}

			p.mut.Lock()
			p.deviationRates = deviations
			p.mut.Unlock()

			return nil
		},
	)

	g.Go(
		func() error {
			medianQueryResponse, err := queryClient.Medians(ctx, &oracletypes.QueryMedians{})
			if err != nil {
				return err
			}

			if len(medianQueryResponse.Medians) == 0 {
				return noMedians
			}

			medians := make(map[string]Median)
			for _, priceStamp := range medianQueryResponse.Medians {
				denom := priceStamp.ExchangeRate.Denom
				amount := priceStamp.ExchangeRate.Amount.Mul(RateFactor).TruncateInt()
				if median, found := medians[denom]; !found {
					medians[denom] = Median{
						Median:    []string{amount.String()},
						Timestamp: strconv.Itoa(int(time.Now().Unix())),
					}
				} else {
					median.Median = append(median.Median, amount.String())
					medians[denom] = median
				}
			}
			p.mut.Lock()
			p.medianRates = medians
			p.mut.Unlock()

			return nil
		},
	)

	return g.Wait()
}

func (p *PriceService) GetPrices(denoms []string) map[string]Price {
	p.mut.Lock()
	exchangeRates := p.exchangeRates
	p.mut.Unlock()

	rates := make(map[string]Price)
	for _, denom := range denoms {
		price, found := exchangeRates[denom]
		if !found {
			continue
		}

		rates[denom] = price
	}

	return rates
}

func (p *PriceService) GetMedians(denoms []string) map[string]Median {
	p.mut.Lock()
	exchangeRates := p.medianRates
	p.mut.Unlock()

	rates := make(map[string]Median)
	for _, denom := range denoms {
		price, found := exchangeRates[denom]
		if !found {
			continue
		}

		rates[denom] = price
	}

	return rates
}

func (p *PriceService) GetDeviations(denoms []string) map[string]Deviation {
	p.mut.Lock()
	exchangeRates := p.deviationRates
	p.mut.Unlock()

	rates := make(map[string]Deviation)
	for _, denom := range denoms {
		price, found := exchangeRates[denom]
		if !found {
			continue
		}

		rates[denom] = price
	}

	return rates
}

// incrementIndex increases index to switch to different query rpc
func (p *PriceService) increment() {
	p.queryRetries += 1
	p.index = (p.index + 1) % len(p.queryRPCS)
	p.logger.Info().Int("rpc index", p.index).Msg("switching query rpc")
}
