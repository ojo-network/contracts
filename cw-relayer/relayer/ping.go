package relayer

import (
	"context"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"

	"github.com/ojo-network/cw-relayer/relayer/client"
)

type UptimeCheck struct {
	logger          zerolog.Logger
	timeoutHeight   int64
	contractAddress string
	relayerAddress  string
	relayerClient   client.RelayerClient
	pingChan        chan types.Msg
}

func NewUptimePing(
	logger zerolog.Logger,
	relayerAddress,
	contractAddress string,
	timeoutHeight int64,
	relayer client.RelayerClient,
	pingChan chan types.Msg,
) *UptimeCheck {
	check := &UptimeCheck{
		contractAddress: contractAddress,
		relayerAddress:  relayerAddress,
		timeoutHeight:   timeoutHeight,
		logger:          logger.With().Str("module", "uptime ping check").Logger(),
		relayerClient:   relayer,
		pingChan:        pingChan,
	}

	return check
}

func (c *UptimeCheck) StartPing(ctx context.Context, duration time.Duration) error {
	c.logger.Info().Msg("uptime check service started ...")
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			err := c.sendPing()
			c.logger.Err(err).Send()
			time.Sleep(duration)
		}
	}
}

func (c *UptimeCheck) sendPing() error {
	msg, err := genPingMsg(c.relayerAddress, c.contractAddress)
	if err != nil {
		return err
	}

	c.logger.Info().Time("ping check", time.Now()).Msg("ping check")
	c.pingChan <- msg

	return nil
}
