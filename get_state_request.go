package pubnub

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/pubnub/go/pnerr"
	"github.com/pubnub/go/utils"
)

const GET_STATE_PATH = "/v2/presence/sub-key/%s/channel/%s/uuid/%s"

var emptyGetStateResp *GetStateResponse

type getStateBuilder struct {
	opts *getStateOpts
}

func newGetStateBuilder(pubnub *PubNub) *getStateBuilder {
	builder := getStateBuilder{
		opts: &getStateOpts{
			pubnub: pubnub,
		},
	}

	return &builder
}

func newGetStateBuilderWithContext(pubnub *PubNub,
	context Context) *getStateBuilder {
	builder := getStateBuilder{
		opts: &getStateOpts{
			pubnub: pubnub,
			ctx:    context,
		},
	}

	return &builder
}

func (b *getStateBuilder) Channels(ch []string) *getStateBuilder {
	b.opts.Channels = ch

	return b
}

func (b *getStateBuilder) ChannelGroups(cg []string) *getStateBuilder {
	b.opts.ChannelGroups = cg

	return b
}

func (b *getStateBuilder) Uuid(uuid string) *getStateBuilder {
	b.opts.Uuid = uuid

	return b
}

func (b *getStateBuilder) Transport(
	tr http.RoundTripper) *getStateBuilder {
	b.opts.Transport = tr

	return b
}

func (b *getStateBuilder) Execute() (
	*GetStateResponse, StatusResponse, error) {
	rawJson, status, err := executeRequest(b.opts)
	if err != nil {
		return emptyGetStateResp, status, err
	}

	return newGetStateResponse(rawJson, status)
}

type getStateOpts struct {
	pubnub *PubNub

	Channels []string

	ChannelGroups []string

	Uuid string

	Transport http.RoundTripper

	ctx Context
}

func (o *getStateOpts) config() Config {
	return *o.pubnub.Config
}

func (o *getStateOpts) client() *http.Client {
	return o.pubnub.GetClient()
}

func (o *getStateOpts) context() Context {
	return o.ctx
}

func (o *getStateOpts) validate() error {
	if o.config().SubscribeKey == "" {
		return newValidationError(o, StrMissingSubKey)
	}

	if len(o.Channels) == 0 && len(o.ChannelGroups) == 0 {
		return newValidationError(o, "Missing Channel or Channel Group")
	}

	return nil
}

func (o *getStateOpts) buildPath() (string, error) {
	var channels []string

	for _, channel := range o.Channels {
		channels = append(channels, utils.PamEncode(channel))
	}

	return fmt.Sprintf(GET_STATE_PATH,
		o.pubnub.Config.SubscribeKey,
		strings.Join(channels, ","),
		utils.UrlEncode(o.pubnub.Config.Uuid)), nil
}

func (o *getStateOpts) buildQuery() (*url.Values, error) {
	q := defaultQuery(o.pubnub.Config.Uuid, o.pubnub.telemetryManager)

	var groups []string

	for _, group := range o.ChannelGroups {
		groups = append(groups, utils.PamEncode(group))
	}

	q.Set("channel-group", strings.Join(groups, ","))

	return q, nil
}

func (o *getStateOpts) buildBody() ([]byte, error) {
	return []byte{}, nil
}

func (o *getStateOpts) httpMethod() string {
	return "GET"
}

func (o *getStateOpts) isAuthRequired() bool {
	return true
}

func (o *getStateOpts) requestTimeout() int {
	return o.pubnub.Config.NonSubscribeRequestTimeout
}

func (o *getStateOpts) connectTimeout() int {
	return o.pubnub.Config.ConnectTimeout
}

func (o *getStateOpts) operationType() OperationType {
	return PNGetStateOperation
}

func (o *getStateOpts) telemetryManager() *TelemetryManager {
	return o.pubnub.telemetryManager
}

type GetStateResponse struct {
	State map[string]interface{}
}

func newGetStateResponse(jsonBytes []byte, status StatusResponse) (
	*GetStateResponse, StatusResponse, error) {

	resp := &GetStateResponse{}

	var value interface{}

	err := json.Unmarshal(jsonBytes, &value)
	if err != nil {
		e := pnerr.NewResponseParsingError("Error unmarshalling response",
			ioutil.NopCloser(bytes.NewBufferString(string(jsonBytes))), err)

		return emptyGetStateResp, status, e
	}

	if v, ok := value.(map[string]interface{}); !ok {
		return emptyGetStateResp, status, errors.New("Response parsing error")
	} else {
		if v["error"] != nil {
			message := ""
			if v["message"] != nil {
				if msg, ok := v["message"].(string); ok {
					message = msg
				}
			}
			return emptyGetStateResp, status, errors.New(message)
		}

		//https://ssp.pubnub.com/v2/presence/sub-key/s/channel/my-channel/uuid/pn-696b6ccf-b473-4b4e-b86e-02ce7eca68cb?pnsdk=PubNub-Go/4.0.0-beta.7&uuid=pn-696b6ccf-b473-4b4e-b86e-02ce7eca68cb
		//
		//https://ps.pubnub.com/v2/presence/sub-key/s/channel/my-channel3,my-channel2,my-channel/uuid/5fef96e6-a64b-4808-8712-3623af768c3b?pnsdk=PubNub-Go/4.0.0-beta.7&uuid=5fef96e6-a64b-4808-8712-3623af768c3b
		//
		m := make(map[string]interface{})
		if v["channel"] != nil {
			if channel, ok2 := v["channel"].(string); ok2 {
				if v["payload"] != nil {
					if val, ok := v["payload"].(interface{}); !ok {
						return emptyGetStateResp, status, errors.New("Response parsing payload")
					} else {
						m[channel] = val
					}
				} else {
					return emptyGetStateResp, status, errors.New("Response parsing channel")
				}
			} else {
				return emptyGetStateResp, status, errors.New("Response parsing channel 2")
			}
		} else {
			if v["payload"] != nil {
				if val, ok := v["payload"].(map[string]interface{}); !ok {
					return emptyGetStateResp, status, errors.New("Response parsing payload 2")
				} else {
					if channels, ok2 := val["channels"].(map[string]interface{}); !ok2 {
						return emptyGetStateResp, status, errors.New("Response parsing channels")
					} else {
						for ch, state := range channels {
							m[ch] = state
						}
					}
				}
			}

		}

		resp.State = m

	}

	return resp, status, nil
}
