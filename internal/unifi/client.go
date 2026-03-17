package unifi

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"sync"
	"time"

	"hikvision-control/internal/common"
)

const offValue = "off"
const autoValue = "auto"

type client struct {
	url       string
	username  string
	password  string
	site      string
	http      *http.Client
	csrfToken string
	apiPrefix string // To handle UniFi OS proxy paths like /proxy/network
	sync.Mutex
}

func NewClient(url, username, password, site string) *client {
	jar, _ := cookiejar.New(nil)
	return &client{
		url:      url,
		username: username,
		password: password,
		site:     site,
		http: &http.Client{
			Timeout: 10 * time.Second,
			Jar:     jar,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

func (c *client) ensureLoggedIn() error {
	c.Lock()
	defer c.Unlock()

	if c.csrfToken != "" {
		return nil
	}

	return c.loginWithLock()
}

func (c *client) Login() error {
	c.Lock()
	defer c.Unlock()

	return c.loginWithLock()
}

func (c *client) loginWithLock() error {
	loginURL := fmt.Sprintf("%s/api/auth/login", c.url)
	payload, _ := json.Marshal(map[string]string{
		"username": c.username,
		"password": c.password,
	})

	req, err := http.NewRequest(http.MethodPost, loginURL, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", c.url)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status: %d", resp.StatusCode)
	}

	// Capture CSRF token for subsequent requests
	c.csrfToken = resp.Header.Get("X-CSRF-Token")

	return nil
}

func (c *client) SetPoeMode(switchMAC string, portIdx int, on bool) error {
	// 1. Ensure we are logged in
	if err := c.ensureLoggedIn(); err != nil {
		return err
	}

	// 2. Get current device state to get ID and current overrides
	device, err := c.GetDeviceInfo(switchMAC)
	if err != nil {
		return err
	}

	newMode := offValue
	if on {
		newMode = autoValue
	}

	// 3. Update port_overrides
	updatedOverrides := make([]common.UnifiPortOverride, 0)
	found := false
	for _, po := range device.PortOverrides {
		if po.PortIdx == portIdx {
			po.PoeMode = newMode
			found = true
		}
		updatedOverrides = append(updatedOverrides, po)
	}

	if !found {
		updatedOverrides = append(updatedOverrides, common.UnifiPortOverride{
			PortIdx: portIdx,
			PoeMode: newMode,
		})
	}

	// 4. PUT the update
	return c.updateDevice(device.DeviceID, updatedOverrides)
}

func (c *client) updateDevice(deviceID string, overrides []common.UnifiPortOverride) error {
	updateURL := fmt.Sprintf("%s%s/api/s/%s/rest/device/%s", c.url, c.apiPrefix, c.site, deviceID)
	payload, _ := json.Marshal(map[string]interface{}{
		"port_overrides": overrides,
	})

	req, _ := http.NewRequest(http.MethodPut, updateURL, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", c.csrfToken)
	req.Header.Set("Referer", c.url)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusUnauthorized {
		// Session might have expired, clear token and retry once
		c.Lock()
		c.csrfToken = ""
		c.Unlock()
		if err := c.ensureLoggedIn(); err != nil {
			return err
		}
		return c.updateDevice(deviceID, overrides)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update port: status %d", resp.StatusCode)
	}

	return nil
}

func (c *client) GetDeviceInfo(mac string) (*common.UnifiDeviceData, error) {
	if err := c.ensureLoggedIn(); err != nil {
		return nil, err
	}

	// First request attempt
	dev, err := c.doGetDeviceInfo(mac)
	if err == nil {
		return dev, nil
	}

	// If 401 Unauthorized, session might have expired
	if err.Error() == "401" {
		c.Lock()
		c.csrfToken = ""
		c.Unlock()
		if err := c.ensureLoggedIn(); err != nil {
			return nil, err
		}
		return c.doGetDeviceInfo(mac)
	}

	// If 404 and we haven't tried the prefix, try with /proxy/network
	if (err.Error() == "404") && c.apiPrefix == "" {
		c.apiPrefix = "/proxy/network"
		dev, err = c.doGetDeviceInfo(mac)
		if err == nil {
			return dev, nil
		}
		// Reset if it still fails so we don't stick with a broken prefix
		c.apiPrefix = ""
	}

	return nil, err
}

func (c *client) doGetDeviceInfo(mac string) (*common.UnifiDeviceData, error) {
	apiURL := fmt.Sprintf("%s%s/api/s/%s/stat/device", c.url, c.apiPrefix, c.site)
	req, _ := http.NewRequest(http.MethodGet, apiURL, nil)
	req.Header.Set("X-CSRF-Token", c.csrfToken)
	req.Header.Set("Referer", c.url)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("401")
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("404")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get devices: status %d", resp.StatusCode)
	}

	// JLS: debug when we need to see the response body because na update renamed the fields
	//bodyBytes, err := io.ReadAll(resp.Body)
	//if err != nil {
	//	return nil, err
	//}
	//
	//ddd := string(bodyBytes)
	//fmt.Println(ddd)

	var devResp common.UnifiDeviceResponse
	err = json.NewDecoder(resp.Body).Decode(&devResp)
	if err != nil {
		return nil, err
	}

	for _, dev := range devResp.Data {
		if dev.MAC == mac {
			return &dev, nil
		}
	}

	return nil, fmt.Errorf("device with MAC %s not found in stat/device list", mac)
}

func (c *client) IsPoeOn(switchMAC string, portIdx int) (bool, error) {
	dev, err := c.GetDeviceInfo(switchMAC)
	if err != nil {
		return false, err
	}

	for _, port := range dev.PortTable {
		if port.PortIdx == portIdx {
			return port.PoeMode != offValue, nil
		}
	}

	return false, fmt.Errorf("port %d not found on switch %s", portIdx, switchMAC)
}
