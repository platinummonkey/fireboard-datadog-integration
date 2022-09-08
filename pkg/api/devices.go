package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	devicesListAPIPath = "/api/v1/devices.json"
	deviceGetAPIPath   = "/api/v1/devices/%s.json"
	deviceTempAPIPath  = "/api/v1/devices/%s/temps.json"
	deviceDriveAPIPath = "/api/v1/devices/%s/drivelog.json"
)

type DriveLogResponse struct {
	DeviceID            int64     `json:"device_id,omitempty"`     // maps back to DevicePropertiesResponse.ID
	DeviceUUID          string    `json:"device_id,omitempty"`     // maps back to DevicePropertiesResponse.UUID
	ModeType            string    `json:"modetype,omitempty"`      // Off or On
	TiedChannel         int64     `json:"tiedchannel,omitempty"`   // the channel that is "tied" not sure what that means
	DrivePercent        float32   `json:"driveper,omitempty"`      // percent [0, 1] of the drive engagement
	SetPoint            float32   `json:"setpoint,omitempty"`      // setpoint temperature a function of degreeType
	Created             time.Time `json:"created,omitempty"`       // the date the log happened
	CreatedMilliseconds int64     `json:"created_ms,omitempty"`    // created in milliseconds since epoch
	UserInitiated       bool      `json:"userinitiated,omitempty"` // 0 if false, 1 if true for user initiated log
	DegreeType          int64     `json:"degreetype,omitempty"`    /// 1 = C, 2 == F
	LidPaused           bool      `json:"lidpaused,omitempty"`     // true if the lid open has caused a pause event
	PowerMode           string    `json:"powermode,omitempty"`     // an enum of the power mode, can be N/A for offline
}

type ChannelAlertConfigResponse struct {
	DeviceID       int64     `json:"device_id,omitempty"`      // the device id for the alert
	ID             int64     `json:"id,omitempty"`             // alert configuration id
	Created        time.Time `json:"created,omitempty"`        // the date alert data was created
	SessionID      int64     `json:"sessionid,omitempty"`      // links back to sessionResponse.ID
	NotifyApp      bool      `json:"notify_app,omitempty"`     // if true it will notify in app
	TemperatureMin float32   `json:"temp_min,omitempty"`       // the minimum temperature alert
	TemperatureMax float32   `json:"temp_max,omitempty"`       // the maximum temperature alert
	Enabled        bool      `json:"enabled,omitempty"`        // set to true if the alert is enabled
	Channel        int64     `json:"channel,omitempty"`        // the channel id this alert is configured for
	NotifySMS      bool      `json:"notify_sms,omitempty"`     // set to true to notify via sms
	TimeStart      time.Time `json:"time_start,omitempty"`     // time for the alert to start being active
	TimeStop       time.Time `json:"time_stop,omitempty"`      // time for the alert to stop being active
	MinutesBuffer  int64     `json:"minutes_buffer,omitempty"` // undocumented
	NotifyEmail    bool      `json:"notify_email,omitempty"`   // if true set to notify via email

}

type ChannelResponse struct {
	SessionID    int64                        `json:"sessionid,omitempty"`     // links back to sessionResponse.ID
	Channel      int64                        `json:"channel,omitempty"`       // the channel id
	ChannelLabel string                       `json:"channel_label,omitempty"` // the name of the channel
	Enabled      bool                         `json:"enabled,omitempty"`       // true if the channel is enabled
	ID           int64                        `json:"id,omitempty"`            // id of the channel
	Created      time.Time                    `json:"created,omitempty"`       // the date channel data was created
	Alerts       []ChannelAlertConfigResponse `json:"alerts,omitempty"`        // alerts configured for this channel
}

type DeviceLog struct {
	InternalIP              string    `json:"internalIP"`     // 1.2.3.4
	AuxillaryPort           string    `json:"auxPort"`        // unknown
	Version                 string    `json:"version"`        // semantic version string
	TxPower                 int       `json:"txpower"`        // dB I think
	Frequency               string    `json:"frequency"`      // 2.4 GHz
	Uptime                  string    `json:"uptime"`         // $hours:$minutes
	SSID                    string    `json:"ssid"`           // string of wifi/bluetooth connection
	MACNIC                  string    `json:"macNIC"`         // mac address of NIC
	CPUUsage                string    `json:"cpuUsage"`       // 66%
	OnboardTemperature      float32   `json:"onboardTemp"`    // float but use degreeType
	SignalLevel             int64     `json:"signallevel"`    // dB I think
	VersionJava             string    `json:"versionJava"`    // semantic version string
	DeviceID                string    `json:"deviceID"`       // uuid of device id
	VoltageBattery          float32   `json:"vBatt"`          // battery voltage
	VersionEspHal           string    `json:"versionEspHal"`  // some awful version string: "HAL: V1R2;AVR: 0.0.14;"
	MemoryUsage             string    `json:"memUsage"`       // 2.7M/4.2M lovely strings
	AccesPointMAC           string    `json:"macAP"`          // wifi access point mac address
	VersionImage            string    `json:"versionImage"`   // semantic version string
	YFBVersion              string    `json:"yfbVersion"`     // semantic version string
	BLEClientMAC            string    `json:"bleClientMAC"`   // BLE Client mac address
	TemperatureFilter       bool      `json:"tempFilter"`     // there is a temp filter enabled
	YFBPower                bool      `json:"yfbPower"`       // does the YFB? have power?
	TimezoneBlueTooth       string    `json:"timeZoneBT"`     // bluetooth timezone configuration America/Chicago
	UtilsVersion            string    `json:"versionUtils"`   // semantic version string
	VoltageBatteryPercent   float32   `json:"vBattPer"`       // Battery Voltage percent
	Contrast                string    `json:"contrast"`       // some [0, ?] range of screen contrast
	LinkQuality             string    `json:"linkquality"`    // 62/100 - could calculate a percent
	DiskUsage               string    `json:"diskUsage"`      // 0.8M/4.0M lovely strings
	PublicIP                string    `json:"publicIP"`       // 1.2.3.4
	NodeVersion             string    `json:"versionNode"`    // node semantic version
	DriveSettings           string    `json:"drivesettings"`  // embedded json..... "{\"p\":0,\"s\":0,\"d\":0,\"ms\":100,\"f\":0,\"l\":1}"
	Date                    time.Time `json:"date"`           // "2022-09-01 00:36:11 UTC"
	Mode                    string    `json:"mode"`           // an enum of sorts - "Managed"
	BoardID                 string    `json:"boardID"`        // board identifier - "GCMABCD12"
	VoltageBatterPercentRaw float32   `json:"vBattPerRaw"`    // not sure but maybe a raw integer value or un-smoothed sample
	Model                   string    `json:"model"`          // model of the device - "YFBX"
	Band                    string    `json:"band"`           // wifi band - "802.11bgn"
	BLESignalLevel          int64     `json:"bleSignalLevel"` // BLE signal level - dB: -93
	YFBModel                string    `json:"yfbModel"`       // some specific model? - "YS640"
	CommercialMode          string    `json:"commercialMode"` // a string representation of "true" or "false"
}

type DevicePropertiesResponse struct {
	ID           int64     `json:"id,omitempty"`            // alternative to unique identifier resource
	UUID         string    `json:"UUID,omitempty"`          // unique identifier resource
	Title        string    `json:"title,omitempty"`         // the name of the fireboard
	Created      time.Time `json:"created,omitempty"`       // the date the fireboard was added to the account
	HardwareID   string    `json:"hardware_id,omitempty"`   // the serial number of the fireboard
	ChannelCount int64     `json:"channel_count,omitempty"` // The device channel count
	Model        string    `json:"model,omitempty"`         // the device model if available
	Active       bool      `json:"active,omitempty"`        // true if the device is actively registered

	LastDriveLog DriveLogResponse `json:"last_drivelog,omitempty"` // last drivelog
	// LatestTemps ...
	DeviceLog DeviceLog `json:"device_log,omitempty"` // the device log

	LastBatteryReading float32 `json:"last_battery_reading,omitempty"` // last battery reading

	Channels    []ChannelResponse `json:"channels,omitempty"`     // channel data
	LastTempLog time.Time         `json:"last_templog,omitempty"` // the date of the last known temperature recorded by this fireboard.

	Version           string `json:"version"`      // version of hardware software version
	FireBoardJVersion string `json:"fbj_version"`  // undocumented fireboard "j" version
	FireBoardNVersion string `json:"fbn_version"`  // undocumented fireboard "n" version
	FireBoardUVersion string `json:"fbu_version"`  // undocumented fireboard "u" version
	ProbeConfig       string `json:"probe_config"` // undocumented probe configuration

}

type ListDevicesResponse []DevicePropertiesResponse

func (a *defaultApiClient) ListDevices() (ListDevicesResponse, error) {
	c := a.getHttpClient()
	urlPath := a.constructURL(devicesListAPIPath)
	ctx, cancel := context.WithTimeout(context.Background(), a.GetTimeout())
	defer cancel()
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		urlPath,
		nil,
	)
	if err != nil {
		return nil, err
	}
	err = SetRequestHeaders(req, a.authStore)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 429 {
		// back off though this is not documented.
		return nil, ErrRateLimited
	} else if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unable to obtain auth token: %v", string(respData))
	}

	var deviceResp ListDevicesResponse
	err = json.Unmarshal(respData, &deviceResp)
	if err != nil {
		return nil, err
	}
	return deviceResp, nil
}

func (a *defaultApiClient) GetDevice(deviceUUID string) (*DevicePropertiesResponse, error) {
	c := a.getHttpClient()
	urlPath := a.constructURL(fmt.Sprintf(deviceGetAPIPath, deviceUUID))
	ctx, cancel := context.WithTimeout(context.Background(), a.GetTimeout())
	defer cancel()
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		urlPath,
		nil,
	)
	if err != nil {
		return nil, err
	}
	err = SetRequestHeaders(req, a.authStore)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 429 {
		// back off though this is not documented.
		return nil, ErrRateLimited
	} else if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unable to obtain auth token: %v", string(respData))
	}

	var deviceResp DevicePropertiesResponse
	err = json.Unmarshal(respData, &deviceResp)
	if err != nil {
		return nil, err
	}
	return &deviceResp, nil
}

func (a *defaultApiClient) GetRealTimeDeviceTemperature(deviceUUID string) (*DevicePropertiesResponse, error) {
	c := a.getHttpClient()
	urlPath := a.constructURL(fmt.Sprintf(deviceTempAPIPath, deviceUUID))
	ctx, cancel := context.WithTimeout(context.Background(), a.GetTimeout())
	defer cancel()
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		urlPath,
		nil,
	)
	if err != nil {
		return nil, err
	}
	err = SetRequestHeaders(req, a.authStore)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 429 {
		// back off though this is not documented.
		return nil, ErrRateLimited
	} else if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unable to obtain auth token: %v", string(respData))
	}

	var deviceResp DevicePropertiesResponse
	err = json.Unmarshal(respData, &deviceResp)
	if err != nil {
		return nil, err
	}
	return &deviceResp, nil
}

func (a *defaultApiClient) GetRealTimeDeviceDriveData(deviceUUID string) (*DevicePropertiesResponse, error) {
	c := a.getHttpClient()
	urlPath := a.constructURL(fmt.Sprintf(deviceDriveAPIPath, deviceUUID))
	ctx, cancel := context.WithTimeout(context.Background(), a.GetTimeout())
	defer cancel()
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		urlPath,
		nil,
	)
	if err != nil {
		return nil, err
	}
	err = SetRequestHeaders(req, a.authStore)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 429 {
		// back off though this is not documented.
		return nil, ErrRateLimited
	} else if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unable to obtain auth token: %v", string(respData))
	}

	var deviceResp DevicePropertiesResponse
	err = json.Unmarshal(respData, &deviceResp)
	if err != nil {
		return nil, err
	}
	return &deviceResp, nil
}
