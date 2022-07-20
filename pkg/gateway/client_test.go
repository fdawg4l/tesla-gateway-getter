package gateway

import (
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	aggregates string = `
{
  "site": {
    "last_communication_time": "2021-03-12T12:20:53.69041677-08:00",
    "instant_power": -13,
    "instant_reactive_power": 37,
    "instant_apparent_power": 39.21734310225516,
    "frequency": 0,
    "energy_exported": 7791.304498571379,
    "energy_imported": 427428.2160958259,
    "instant_average_voltage": 212.90368526636638,
    "instant_total_current": 9.892,
    "i_a_current": 0,
    "i_b_current": 0,
    "i_c_current": 0,
    "last_phase_voltage_communication_time": "0001-01-01T00:00:00Z",
    "last_phase_power_communication_time": "0001-01-01T00:00:00Z",
    "timeout": 1500000000
  },
  "battery": {
    "last_communication_time": "2021-03-12T12:20:53.690236531-08:00",
    "instant_power": -3260,
    "instant_reactive_power": 20,
    "instant_apparent_power": 3260.0613491159947,
    "frequency": 60.010999999999996,
    "energy_exported": 201230,
    "energy_imported": 245960,
    "instant_average_voltage": 246.4,
    "instant_total_current": 70,
    "i_a_current": 0,
    "i_b_current": 0,
    "i_c_current": 0,
    "last_phase_voltage_communication_time": "0001-01-01T00:00:00Z",
    "last_phase_power_communication_time": "0001-01-01T00:00:00Z",
    "timeout": 1500000000
  },
  "load": {
    "last_communication_time": "2021-03-12T12:20:53.690236531-08:00",
    "instant_power": 764.25,
    "instant_reactive_power": 2.75,
    "instant_apparent_power": 764.2549476450905,
    "frequency": 0,
    "energy_exported": 0,
    "energy_imported": 720901.1981976463,
    "instant_average_voltage": 212.90368526636638,
    "instant_total_current": 3.5896513441927396,
    "i_a_current": 0,
    "i_b_current": 0,
    "i_c_current": 0,
    "last_phase_voltage_communication_time": "0001-01-01T00:00:00Z",
    "last_phase_power_communication_time": "0001-01-01T00:00:00Z",
    "timeout": 1500000000
  },
  "solar": {
    "last_communication_time": "2021-03-12T12:20:53.690663131-08:00",
    "instant_power": 4022,
    "instant_reactive_power": -28,
    "instant_apparent_power": 4022.0974627674054,
    "frequency": 0,
    "energy_exported": 347722.9480106806,
    "energy_imported": 1728.6614102888489,
    "instant_average_voltage": 212.869044250215,
    "instant_total_current": 16.557,
    "i_a_current": 0,
    "i_b_current": 0,
    "i_c_current": 0,
    "last_phase_voltage_communication_time": "0001-01-01T00:00:00Z",
    "last_phase_power_communication_time": "0001-01-01T00:00:00Z",
    "timeout": 1500000000
  }
}
`

	percentage string = `{
  "percentage": 60.3277636547622
}
`
)

func TestUnmarshal(t *testing.T) {
	a := new(Aggregates)
	if !assert.NoError(t, json.Unmarshal([]byte(aggregates), a)) {
		return
	}

	require.Equal(t, a.Values["battery_instant_reactive_power"].(float64), float64(20))

	b := new(SOE)
	require.NoError(t, json.Unmarshal([]byte(percentage), b))

	require.Equal(t, b.Percentage, float64(60.3277636547622))

	reflector := NewReflector()
	srv := httptest.NewServer(reflector)
	defer srv.Close()

	reflector.Reflect(a.Values, b.Percentage)

	resp, err := srv.Client().Get(srv.URL)
	require.NoError(t, err)

	var out map[string]json.RawMessage
	dec := json.NewDecoder(resp.Body)
	require.NoError(t, dec.Decode(&out))

	c := new(Aggregates)
	require.NoError(t, json.Unmarshal(out["inverter"], &c.Values))
	// dirty hack because we're not marshalling and unmarshalling
	// correctly, and I don't want to change the database.
	require.Equal(t, c.Values["battery_instant_reactive_power"].(float64), float64(20))

	d := new(SOE)
	require.NoError(t, json.Unmarshal(out["battery"], &d.Percentage))
	require.Equal(t, d, b)

	//out, _ := ioutil.ReadAll(resp.Body)
	//fmt.Printf("%s\n", string(out))
}

func TestLoginRequest(t *testing.T) {
	l := NewLogin("foo", "bar")
	u, _ := url.Parse("http://nowhere")
	req, err := l.Request(u)
	if assert.NoError(t, err) {
		return
	}

	out := new(Login)
	d := json.NewDecoder(req.Body)
	if assert.NoError(t, d.Decode(out)) {
		return
	}

	if !assert.Equal(t, l, out) {
		return
	}
}
