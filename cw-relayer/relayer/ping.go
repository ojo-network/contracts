package relayer

import (
	"context"
	"time"

	"github.com/rs/zerolog"

	"github.com/ojo-network/cw-relayer/relayer/client"
)

type UptimeCheck struct {
	logger          zerolog.Logger
	timeoutHeight   int64
	contractAddress string
	relayerAddress  string
	relayerClient   client.RelayerClient
}

func NewUptimePing(
	logger zerolog.Logger,
	relayerAddress,
	contractAddress string,
	timeoutHeight int64,
	relayer client.RelayerClient,
) *UptimeCheck {
	check := &UptimeCheck{
		contractAddress: contractAddress,
		relayerAddress:  relayerAddress,
		timeoutHeight:   timeoutHeight,
		logger:          logger.With().Str("module", "uptime ping check").Logger(),
		relayerClient:   relayer,
	}

	return check
}

func (c *UptimeCheck) StartPing(ctx context.Context, duration time.Duration) error {
	c.logger.Info().Msg("uptime check service started ...")
	ticker := time.NewTicker(duration)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			err := c.sendPing()
			c.logger.Err(err).Send()
		}
	}
}

func (c *UptimeCheck) sendPing() error {
	msg, err := genPingMsg(c.relayerAddress, c.contractAddress)
	if err != nil {
		return err
	}

	c.logger.Info().Time("ping check", time.Now()).Msg("ping check")
	return c.relayerClient.BroadcastTx(c.timeoutHeight, msg)
}
