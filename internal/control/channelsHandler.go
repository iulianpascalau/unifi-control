package control

import (
	"errors"
	"fmt"

	"unifi-control/internal/common"
	"unifi-control/internal/config"

	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("channels-handler")

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
	log.Debug("GetPortIDs called", "count", len(h.portIDs))
	return h.portIDs
}

func (h *channelsHandler) GetPort(portID string) common.PortStatus {
	cfg, ok := h.portsAsMap[portID]
	if !ok {
		log.Error("port not found in configuration", "portID", portID)
		return common.PortStatus{
			Name:   unknownName,
			PortID: portID,
			Active: false,
			Error:  fmt.Sprintf("port id %s not found", portID),
		}
	}

	switchMAC, switchName, portIdx, err := h.resolveCameraLocation(cfg)
	if err != nil {
		log.Error("failed to resolve camera location", "name", cfg.Name, "cameraMAC", cfg.CameraMAC, "error", err)
		return common.PortStatus{
			Name:   cfg.Name,
			PortID: portID,
			Active: false,
			Error:  fmt.Sprintf("discovery error: %v", err),
		}
	}

	device, err := h.unifiClient.GetDeviceInfo(switchMAC)
	if err != nil {
		log.Error("failed to get device info from UniFi", "name", cfg.Name, "switchMAC", switchMAC, "error", err)
		return common.PortStatus{
			Name:   cfg.Name,
			PortID: portID,
			Active: false,
			Error:  fmt.Sprintf("unifi error: %v", err),
		}
	}

	for _, port := range device.PortTable {
		if port.PortIdx == portIdx {
			log.Debug("successfully fetched port status", "name", cfg.Name, "switchMAC", switchMAC, "portIdx", portIdx, "active", port.PoeMode != "off")
			return common.PortStatus{
				Name:       cfg.Name,
				PortID:     portID,
				Active:     port.PoeMode != "off",
				PoePower:   port.PoePower,
				PoeCurrent: port.PoeCurrent,
				PoeVoltage: port.PoeVoltage,
				PoeClass:   port.PoeClass,
				SwitchMAC:  switchMAC,
				SwitchName: switchName,
				PortIdx:    portIdx,
				Error:      "",
			}
		}
	}

	log.Error("port index not found on switch", "name", cfg.Name, "switchMAC", switchMAC, "portIdx", portIdx)
	return common.PortStatus{
		Name:   cfg.Name,
		PortID: portID,
		Active: false,
		Error:  fmt.Sprintf("port %d not found on switch %s", portIdx, switchMAC),
	}
}

// Set updates the channel IP addresses on a specified channel with a provided bool value
func (h *channelsHandler) Set(portID string, active bool) error {
	cfg, ok := h.portsAsMap[portID]
	if !ok {
		log.Error("port not found in configuration for Set", "portID", portID)
		return fmt.Errorf("port %s not found", portID)
	}

	switchMAC, _, portIdx, err := h.resolveCameraLocation(cfg)
	if err != nil {
		log.Error("failed to resolve camera location for Set", "name", cfg.Name, "cameraMAC", cfg.CameraMAC, "error", err)
		return err
	}

	log.Debug("updating port PoE state", "name", cfg.Name, "switchMAC", switchMAC, "portIdx", portIdx, "active", active)
	return h.unifiClient.SetPoeMode(switchMAC, portIdx, active)
}

func (h *channelsHandler) resolveCameraLocation(cfg config.PortConfig) (string, string, int, error) {
	if cfg.CameraMAC == "" {
		return cfg.SwitchMAC, "Static", cfg.SwitchPort, nil
	}

	devices, err := h.unifiClient.GetAllDevices()
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to fetch devices for discovery: %w", err)
	}

	for _, dev := range devices {
		for _, port := range dev.PortTable {
			if port.LastConnection.MAC == cfg.CameraMAC {
				name := dev.Name
				if name == "" {
					name = dev.MAC
				}
				return dev.MAC, name, port.PortIdx, nil
			}
		}
	}

	return "", "", 0, fmt.Errorf("camera with MAC %s not found on any switch port", cfg.CameraMAC)
}
