package relayer

import (
	"context"
	"github.com/ojo-network/cw-relayer/relayer/client"
	"github.com/rs/zerolog"
	"time"
)

type UptimeCheck struct {
	logger          zerolog.Logger
	contractAddress string
	relayerAddress  string
	relayerClient   client.RelayerClient
}

func NewUptimePing(
	logger zerolog.Logger,
	relayerAddress,
	contractAddress string,
	relayer client.RelayerClient,
) *UptimeCheck {
	check := &UptimeCheck{
		contractAddress: contractAddress,
		relayerAddress:  relayerAddress,
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
	ping := Ping{Ping: struct{}{}}
	msg, err := genMsg(c.relayerAddress, c.contractAddress, ping)
	if err != nil {
		return err
	}

	return c.relayerClient.BroadcastTx(0, msg)
}
