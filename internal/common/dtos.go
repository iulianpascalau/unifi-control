package common

type PortStatus struct {
	Name       string `json:"name"`
	PortID     string `json:"port-id"`
	Active     bool   `json:"status"`
	PoePower   string `json:"poe_power,omitempty"`
	PoeCurrent string `json:"poe_current,omitempty"`
	PoeVoltage string `json:"poe_voltage,omitempty"`
	PoeClass   string `json:"poe_class,omitempty"`
	SwitchMAC  string `json:"switch_mac,omitempty"`
	SwitchName string `json:"switch_name,omitempty"`
	PortIdx    int    `json:"port_idx,omitempty"`
	Error      string `json:"error"`
}

type UnifiLastConnection struct {
	MAC      string `json:"mac"`
	LastSeen int64  `json:"last_seen"`
}

type UnifiPortStatus struct {
	PortIdx        int                 `json:"port_idx"`
	Up             bool                `json:"up"`
	PoeMode        string              `json:"poe_mode"`
	PoePower       string              `json:"poe_power"`
	PoeCurrent     string              `json:"poe_current"`
	PoeVoltage     string              `json:"poe_voltage"`
	PoeClass       string              `json:"poe_class"`
	LastConnection UnifiLastConnection `json:"last_connection"`
}

type UnifiDeviceResponse struct {
	Data []UnifiDeviceData `json:"data"`
}

type UnifiDeviceData struct {
	MAC           string                   `json:"mac"`
	Name          string                   `json:"name"`
	DeviceID      string                   `json:"_id"`
	PortOverrides []map[string]interface{} `json:"port_overrides"`
	PortTable     []UnifiPortStatus        `json:"port_table"`
}
