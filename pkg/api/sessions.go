package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	sessionsListAPIPath     = "/api/v1/sessions.json"
	sessionsGetAPIPath      = "/api/v1/sessions/%d.json?drive=1"
	sessionChartDataAPIPath = "/api/v1/sessions/%d/chart.json?drive=1"
)

type SessionListResponse struct {
	ID          int64     `json:"id,omitempty"`          // unique identifier resource
	Title       string    `json:"title,omitempty"`       // the name of the session
	Duration    string    `json:"duration,omitempty"`    // the duration fo the total session duration of this session (string form in api 5 hours, 30 minutes)
	Created     time.Time `json:"created,omitempty"`     // the date the session was created in the fireboard cloud
	StartTime   time.Time `json:"start_time,omitempty"`  // the configurable start time for the session
	EndTime     time.Time `json:"end_time,omitempty"`    // the configurable end time for the session
	Description string    `json:"description,omitempty"` // a string containing notes entered by the user pertaining to the session
	Shared      bool      `json:"shared,omitempty"`      // if this sessions was shared in-app via social media
	ShareKey    string    `json:"share_key,omitempty"`   // the shared key for this session
	DeviceIDs   []string  `json:"device_ids,omitempty"`  // array of device ids
}

type SessionsListResponse []SessionListResponse

func (a *defaultApiClient) ListAllSessions() (SessionsListResponse, error) {
	c := a.getHttpClient()
	urlPath := a.constructURL(sessionsListAPIPath)
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

	var sResp SessionsListResponse
	err = json.Unmarshal(respData, &sResp)
	if err != nil {
		return nil, err
	}
	return sResp, nil
}

type UserProfileResponse struct {
	Company          string    `json:"company,omitempty"`           // company of the user
	AlertSMS         string    `json:"alert_sms,omitempty"`         // sms phone number of the user if alerts configured
	AlertEmails      string    `json:"alert_emails,omitempty"`      // csv of alert emails if configured
	NotificationTone string    `json:"notification_tone,omitempty"` // notification default, an eum
	User             int64     `json:"user,omitempty"`              // matches the OwnerResponse.ID
	Picture          string    `json:"picture,omitempty"`           // the configured user profile picture
	LastTempLog      time.Time `json:"last_templog,omitempty"`      // last time this user looked at the temp log
	CommercialUser   bool      `json:"commercial_user,omitempty"`   // if this is a commerical user this is true
	Beta             bool      `json:"beta,omitempty"`              // true if this user is enrolled in the beta program
	TOS              bool      `json:"tos,omitempty"`               // true if the user agreed to the terms of service.
}

type OwnerResponse struct {
	Username    string              `json:"username,omitempty"`    // username of the owner
	Email       string              `json:"email,omitempty"`       // email of the owner
	FirstName   string              `json:"first_name,omitempty"`  // first name of the user
	LastName    string              `json:"last_name,omitempty"`   // first name of the user
	ID          int64               `json:"id,omitempty"`          // unique id of the user
	UserProfile UserProfileResponse `json:"userprofile,omitempty"` // user profile information on the owner
}

type SessionGetResponse struct {
	ID          int64     `json:"id,omitempty"`          // unique identifier resource
	Title       string    `json:"title,omitempty"`       // the name of the session
	Description string    `json:"description,omitempty"` // a string containing notes entered by the user pertaining to the session
	Duration    string    `json:"duration,omitempty"`    // the duration fo the total sesstion duration of this session (string form in api 5 hours, 30 minutes)
	Created     time.Time `json:"created,omitempty"`     // the date the session was created in the fireboard cloud
	StartTime   time.Time `json:"start_time,omitempty"`  // the configurable start time for the session
	EndTime     time.Time `json:"end_time,omitempty"`    // the configurable end time for the session
	LastActive  time.Time `json:"last_active,omitempty"` // the last active time of this session

	Shared   bool   `json:"shared,omitempty"`    // if this sessions was shared in-app via social media
	ShareKey string `json:"share_key,omitempty"` // the shared key for this session
	Drive    bool   `json:"drive,omitempty"`     // if the drive was enabled on this session

	CanManage bool `json:"can_manage,omitempty"` // if true this is a managable session

	DeviceIDs []string                   `json:"device_ids,omitempty"` // array of device ids
	Devices   []DevicePropertiesResponse `json:"devices,omitempty"`    // device information for this session

	Owner OwnerResponse `json:"owner,omitempty"` // owner information
}

func (a *defaultApiClient) GetSession(sessionID int64) (*SessionGetResponse, error) {
	c := a.getHttpClient()
	urlPath := a.constructURL(fmt.Sprintf(sessionsGetAPIPath, sessionID))
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

	var sResp SessionGetResponse
	err = json.Unmarshal(respData, &sResp)
	if err != nil {
		return nil, err
	}
	return &sResp, nil
}

type SessionChartObject struct {
	ChannelID  json.Number `json:"channel_id,omitempty"` // the channel id. ${type}_${uuid} or an integer
	DegreeType int64       `json:"degreetype,omitempty"` // 1 = C, 2 = F
	Label      int64       `json:"label,omitempty"`      // the channel label
	Device     string      `json:"device,omitempty"`     // the device UUID
	X          []int64     `json:"x,omitempty"`          // the timestamps in epoch seconds for the data
	Y          []float32   `json:"y,omitempty"`          // the values in degreeType if a temperature
}

// return the channel type
func (s SessionChartObject) ChannelType() string {
	if _, err := s.ChannelID.Int64(); err == nil {
		return "temperature"
	}

	return strings.SplitN(s.ChannelID.String(), "_", 2)[0]
}

type SessionChartResponse []SessionChartObject

func (a *defaultApiClient) GetSessionChartData(sessionID int64) (SessionChartResponse, error) {
	c := a.getHttpClient()
	urlPath := a.constructURL(fmt.Sprintf(sessionChartDataAPIPath, sessionID))
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

	var sResp SessionChartResponse
	err = json.Unmarshal(respData, &sResp)
	if err != nil {
		return nil, err
	}
	return sResp, nil
}
