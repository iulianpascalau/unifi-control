package control

import (
	"bytes"
	"errors"
	"fmt"

	"hikvision-control/internal/common"
	"hikvision-control/internal/config"
	"hikvision-control/internal/unifi"

	"github.com/multiversx/mx-chain-core-go/core/check"
)

const unknownName = "unknown"

type channelsHandler struct {
	channelIDs    []string
	channelsAsMap map[string]config.ChannelConfig
	hikHandler    HikvisionHandler
	unifiClient   *unifi.Client
}

func NewChannelsHandler(cfg config.Config, unifiClient *unifi.Client, hikHandler HikvisionHandler) (*channelsHandler, error) {
	channelIDs, err := processChannels(cfg.Channels)
	if err != nil {
		return nil, err
	}

	if check.IfNilReflect(hikHandler) {
		return nil, errors.New("hikvision handler is nil")
	}

	return &channelsHandler{
		channelsAsMap: sliceToMap(cfg.Channels),
		channelIDs:    channelIDs,
		hikHandler:    hikHandler,
		unifiClient:   unifiClient,
	}, nil
}

func processChannels(channels []config.ChannelConfig) ([]string, error) {
	m := make(map[string]int)

	channelIDs := make([]string, 0, len(channels))
	for _, ch := range channels {
		m[ch.Channel]++
		channelIDs = append(channelIDs, ch.Channel)
	}

	for ch, count := range m {
		if count > 1 {
			return nil, fmt.Errorf("channel %s was defined more than once", ch)
		}
	}

	return channelIDs, nil
}

func sliceToMap(channels []config.ChannelConfig) map[string]config.ChannelConfig {
	m := make(map[string]config.ChannelConfig)
	for _, ch := range channels {
		m[ch.Channel] = ch
	}

	return m
}

func (h *channelsHandler) GetChannels() []string {
	return h.channelIDs
}

func (h *channelsHandler) GetChannel(channel string) common.ChannelStatus {
	cfg, ok := h.channelsAsMap[channel]
	if !ok {
		return common.ChannelStatus{
			Name:    unknownName,
			Channel: channel,
			Active:  false,
			Error:   fmt.Sprintf("channel %s not found", channel),
		}
	}

	poeOn, err := h.unifiClient.IsPoeOn(cfg.SwitchMAC, cfg.Port)
	if err != nil {
		return common.ChannelStatus{
			Name:    cfg.Name,
			Channel: channel,
			Active:  false,
			Error:   fmt.Sprintf("unifi error: %v", err),
		}
	}

	return common.ChannelStatus{
		Name:    cfg.Name,
		Channel: channel,
		Active:  poeOn,
		Error:   "",
	}
}

// Set updates the channel IP addresses on a specified channel with a provided bool value
func (h *channelsHandler) Set(channel string, active bool) error {
	cfg, ok := h.channelsAsMap[channel]
	if !ok {
		return fmt.Errorf("channel %s not found", channel)
	}

	return h.unifiClient.SetPoeMode(cfg.SwitchMAC, cfg.Port, active)
}
