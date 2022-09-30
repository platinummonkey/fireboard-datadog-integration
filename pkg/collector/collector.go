package collector

import (
	"fmt"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"

	"github.com/platinummonkey/fireboard-datadog-integration/pkg/api"
)

type collector struct {
	client api.APIClient
	stat   statsd.ClientInterface
	tags   []string
}

func NewCollector(client api.APIClient, stat statsd.ClientInterface, tags []string) *collector {
	if client == nil {
		client = api.NewDefaultAPIClient()
	}
	return &collector{
		client: client,
		stat:   stat,
		tags:   tags,
	}
}

func (c *collector) Authenticate(username, password string) error {
	_, err := c.client.GetAuthToken(username, password)
	return err
}

func (c *collector) Collect(cutoffDate time.Time, stat statsd.ClientInterface) error {
	devices, err := c.client.ListDevices()
	if err != nil {
		c.stat.Incr("fireboard.devices.errors", append(c.tags, "func:devicesList"), 1.0)
		return err
	}
	c.stat.Count("fireboard.devices", int64(len(devices)), c.tags, 1)
	for _, device := range devices {
		if device.Active {
			uuidTag := "uuid:" + device.UUID
			tags := append(c.tags, uuidTag)
			c.stat.Incr("fireboard.devices.active", tags, 1.0)
			c.stat.Gauge("fireboard.devices.link_quality", device.DeviceLog.LinkQualityPercent(), append(tags, "ssid:"+device.DeviceLog.SSID), 1.0)
			c.stat.Gauge("fireboard.devices.disk_usage_percent", device.DeviceLog.DiskUsagePercent(), tags, 1.0)
			c.stat.Gauge("fireboard.devices.memory_usage_percent", device.DeviceLog.MemoryUsagePercent(), tags, 1.0)
			c.stat.Gauge("fireboard.devices.cpu_usage_percent", device.DeviceLog.CPUPercent(), tags, 1.0)
			/*
				driveData, err := c.client.GetRealTimeDeviceDriveData(device.UUID)
				if err != nil {
					c.stat.Incr("fireboard.devices.errors", append(c.tags, uuidTag, "func:devicesGetRealtimeDeviceDriveData"), 1.0)
					return err
				}
				// TODO: report drive data

				tempData, err := c.client.GetRealTimeDeviceTemperature(device.UUID)
				if err != nil {
					c.stat.Incr("fireboard.devices.errors", append(c.tags, uuidTag, "func:devicesGetRealtimeTemperatureData"), 1.0)
					return err
				}
				// TODO: report temp data
			*/
		}
		// do something with cutoff date
	}

	sessions, err := c.client.ListAllSessions()
	c.stat.Count("fireboard.sessions", int64(len(sessions)), c.tags, 1.0)
	if err != nil {
		c.stat.Incr("fireboard.sessions.errors", append(c.tags, "func:sessionsList"), 1.0)
		return err
	}
	for _, session := range sessions {
		active := session.EndTime.After()
		sessionIDTag := fmt.Sprintf("sessionID:%d", session.ID)
		tags := append(c.tags, sessionIDTag)
		if active {
			c.stat.Incr("fireboard.sessions.active", tags, 1.0)
		}
		if session.EndTime.After(cutoffDate) {
			// do something with
			chartDataForSession, err := c.client.GetSessionChartData(session.ID)
			if err != nil {
				c.stat.Incr("fireboard.devices.errors", append(tags, "func:sessionsGetChartData"), 1.0)
				return err
			}
			// TODO report chart data
			for _, sensor := range chartDataForSession {
				sensorTags := append(tags, "label:" + sensor.Label, "device_id:" + sensor.Device)
				conversion := unity
				if sensor.DegreeType == 2 {
					conversion = fToC
				}
				// fixme for non temp sensors ^^^
				for i := 0; i < len(sensor.X); i++ {
					d := time.Unix(sensor.X[i], 0)
					if d.After(time.Now().Add(time.Minute * -30)) {
						// ignore all other data it's too old to ingest
						v := sensor.Y[1]
						c.stat.Gauge("fireboard.sessions.chart", conversion(v))
					}
				}
			}
		}
	}
}


func unity(v float32) float32 {
	return v
}

func fToC(v float32) float32 {
	return (v - 32) * 5.0/9.0
}