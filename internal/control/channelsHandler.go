package control

import (
	"errors"
	"fmt"

	"hikvision-control/internal/common"
	"hikvision-control/internal/config"

	"github.com/multiversx/mx-chain-core-go/core/check"
)

const unknownName = "unknown"

type channelsHandler struct {
	portIDs     []string
	portsAsMap  map[string]config.PortConfig
	unifiClient UnifiHandler
}

func NewChannelsHandler(cfg config.Config, unifiClient UnifiHandler) (*channelsHandler, error) {
	if check.IfNilReflect(unifiClient) {
		return nil, errors.New("unifi handler is nil")
	}

	portIDs, mPorts := processPorts(cfg.Ports)

	return &channelsHandler{
		portsAsMap:  mPorts,
		portIDs:     portIDs,
		unifiClient: unifiClient,
	}, nil
}

func processPorts(channels []config.PortConfig) ([]string, map[string]config.PortConfig) {
	mPorts := make(map[string]config.PortConfig)

	channelIDs := make([]string, 0, len(channels))
	for i, ch := range channels {
		portID := fmt.Sprintf("%d", i)
		channelIDs = append(channelIDs, portID)
		mPorts[portID] = ch
	}

	return channelIDs, mPorts
}

func (h *channelsHandler) GetPortIDs() []string {
	return h.portIDs
}

func (h *channelsHandler) GetPort(portID string) common.PortStatus {
	cfg, ok := h.portsAsMap[portID]
	if !ok {
		return common.PortStatus{
			Name:   unknownName,
			PortID: portID,
			Active: false,
			Error:  fmt.Sprintf("port id %s not found", portID),
		}
	}

	device, err := h.unifiClient.GetDeviceInfo(cfg.SwitchMAC)
	if err != nil {
		return common.PortStatus{
			Name:   cfg.Name,
			PortID: portID,
			Active: false,
			Error:  fmt.Sprintf("unifi error: %v", err),
		}
	}

	for _, port := range device.PortTable {
		if port.PortIdx == cfg.Port {
			return common.PortStatus{
				Name:       cfg.Name,
				PortID:     portID,
				Active:     port.PoeMode != "off",
				PoePower:   port.PoePower,
				PoeCurrent: port.PoeCurrent,
				PoeVoltage: port.PoeVoltage,
				PoeClass:   port.PoeClass,
				Error:      "",
			}
		}
	}

	return common.PortStatus{
		Name:   cfg.Name,
		PortID: portID,
		Active: false,
		Error:  fmt.Sprintf("port %d not found on switch %s", cfg.Port, cfg.SwitchMAC),
	}
}

// Set updates the channel IP addresses on a specified channel with a provided bool value
func (h *channelsHandler) Set(portID string, active bool) error {
	cfg, ok := h.portsAsMap[portID]
	if !ok {
		return fmt.Errorf("port %s not found", portID)
	}

	return h.unifiClient.SetPoeMode(cfg.SwitchMAC, cfg.Port, active)
}
