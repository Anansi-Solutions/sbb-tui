// Package model defines the data types decoded from the search.ch timetable API.
package model

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"
)

// SwissLocation is the Europe/Zurich time zone, with a fixed-offset
// fallback when the tzdata lookup fails. The fallback does not handle
// the CET-CEST transition; main.go embeds time/tzdata to keep that
// path off the hot road.
var SwissLocation = func() *time.Location {
	loc, err := time.LoadLocation("Europe/Zurich")
	if err != nil {
		loc = time.FixedZone("CET", 1*60*60)
	}
	return loc
}()

// Timestamp is a time.Time decoded from the API's "2006-01-02 15:04:05"
// format, which is expressed in Swiss local time.
type Timestamp struct {
	time.Time
}

// Sub returns the duration between two Timestamps.
func (t Timestamp) Sub(other Timestamp) time.Duration {
	return t.Time.Sub(other.Time)
}

// Local returns the timestamp in Swiss time, overriding time.Time.Local.
func (t Timestamp) Local() time.Time {
	return t.In(SwissLocation)
}

// UnmarshalJSON parses the API's "2006-01-02 15:04:05" format in Swiss local time.
func (t *Timestamp) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "" {
		return nil
	}
	parsed, err := time.ParseInLocation("2006-01-02 15:04:05", s, SwissLocation)
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

// Delay is a delay decoded from the API's string format: "+N" for N
// minutes of delay, or "X" when the stop/leg is cancelled.
type Delay struct {
	Minutes   int
	Cancelled bool
}

// UnmarshalJSON tolerantly parses "+0", "+5", "X", "" and null.
func (d *Delay) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "" || s == "null" {
		return nil
	}
	if strings.EqualFold(s, "X") {
		d.Cancelled = true
		return nil
	}
	if n, err := strconv.Atoi(strings.TrimPrefix(s, "+")); err == nil {
		d.Minutes = n
	}
	return nil
}

// Occupancy is the API's two-digit occupancy code: the first digit is the
// first-class load, the second the second-class load (1=low, 2=medium,
// 3=high, 0/absent=unknown).
type Occupancy string

// digit returns the numeric value of the occupancy digit at index i, or 0.
func (o Occupancy) digit(i int) int {
	if len(o) <= i || o[i] < '0' || o[i] > '9' {
		return 0
	}
	return int(o[i] - '0')
}

// FirstClass returns the first-class occupancy level (0=unknown, 1=low, 2=medium, 3=high).
func (o Occupancy) FirstClass() int { return o.digit(0) }

// SecondClass returns the second-class occupancy level (0=unknown, 1=low, 2=medium, 3=high).
func (o Occupancy) SecondClass() int { return o.digit(1) }

// Stop is an intermediate halt within a Leg.
type Stop struct {
	Name      string    `json:"name"`
	StopID    string    `json:"stopid"`
	Arrival   Timestamp `json:"arrival"`
	Departure Timestamp `json:"departure"`
	DepDelay  Delay     `json:"dep_delay"`
	Occupancy Occupancy `json:"occupancy"`
	Lat       float64   `json:"lat"`
	Lon       float64   `json:"lon"`
}

// Exit describes where (and when) a leg is left.
type Exit struct {
	Name      string    `json:"name"`
	SBBName   string    `json:"sbb_name"`
	StopID    string    `json:"stopid"`
	Arrival   Timestamp `json:"arrival"`
	ArrDelay  Delay     `json:"arr_delay"`
	Track     string    `json:"track"`
	WaitTime  int       `json:"waittime"` // seconds until the next departure
	IsAddress bool      `json:"isaddress"`
	Lat       float64   `json:"lat"`
	Lon       float64   `json:"lon"`
}

// Leg is one node of a connection: the stop where you board (or start
// walking) plus the vehicle ridden and the Exit where you get off.
// The final leg of a connection is a bare arrival node: it has a Name
// and Arrival but no Exit, vehicle or walk attached.
type Leg struct {
	Name        string    `json:"name"`
	SBBName     string    `json:"sbb_name"`
	StopID      string    `json:"stopid"`
	TripID      string    `json:"tripid"`
	Arrival     Timestamp `json:"arrival"`   // arrival at this node
	Departure   Timestamp `json:"departure"` // departure from this node
	DepDelay    Delay     `json:"dep_delay"`
	Type        string    `json:"type"`      // walk, bus, tram, strain, express_train, ship, ...
	TypeName    string    `json:"type_name"` // human-readable vehicle kind
	Line        string    `json:"line"`      // e.g. "S13", "IC 1", "17"
	Category    string    `json:"*G"`        // e.g. "IC"
	LineNumber  string    `json:"*L"`        // e.g. "1"
	Terminal    string    `json:"terminal"`  // direction the vehicle is headed
	Operator    string    `json:"operator"`
	FgColor     string    `json:"fgcolor"` // line color, hex without '#'
	BgColor     string    `json:"bgcolor"` // line color, hex without '#'
	Track       string    `json:"track"`   // "13"; a trailing '!' marks a platform change
	Occupancy   Occupancy `json:"occupancy"`
	RunningTime int       `json:"runningtime"` // seconds on board / walking
	WaitTime    int       `json:"waittime"`    // seconds waited before departure
	IsAddress   bool      `json:"isaddress"`
	Lat         float64   `json:"lat"`
	Lon         float64   `json:"lon"`
	Exit        *Exit     `json:"exit"`
	Stops       []Stop    `json:"stops"`
}

// IsWalk reports whether the leg is a walking transfer.
func (l Leg) IsWalk() bool { return l.Type == "walk" }

// OperatorName returns the operator with the API's stray internal
// whitespace collapsed (e.g. "VBZ    F" -> "VBZ F").
func (l Leg) OperatorName() string {
	return strings.Join(strings.Fields(l.Operator), " ")
}

// IsVehicle reports whether the leg is ridden on a transport line.
func (l Leg) IsVehicle() bool { return l.Line != "" && !l.IsWalk() }

// Connection is one full route returned by the API.
type Connection struct {
	From      string    `json:"from"`
	To        string    `json:"to"`
	Departure Timestamp `json:"departure"`
	Arrival   Timestamp `json:"arrival"`
	Duration  int       `json:"duration"` // seconds
	DepDelay  Delay     `json:"dep_delay"`
	ArrDelay  Delay     `json:"arr_delay"`
	Occupancy Occupancy `json:"occupancy"`
	Legs      []Leg     `json:"legs"`

	// Disruptions is kept raw: the API returns an object keyed by
	// disruption ID, but historically also an empty array.
	Disruptions json.RawMessage `json:"disruptions"`
}

// DisruptionTexts are the alert texts of a disruption in one verbosity level.
type DisruptionTexts struct {
	Summary     string `json:"summary"`
	Duration    string `json:"duration"`
	Description string `json:"description"`
	Consequence string `json:"consequence"`
}

// Disruption is one service alert attached to a connection.
type Disruption struct {
	ID    string                     `json:"id"`
	Texts map[string]DisruptionTexts `json:"texts"`
}

// Text returns the shortest text variant available (S, then M, then L).
func (d Disruption) Text() DisruptionTexts {
	for _, k := range []string{"S", "M", "L"} {
		if t, ok := d.Texts[k]; ok {
			return t
		}
	}
	return DisruptionTexts{}
}

// DisruptionList decodes the raw disruptions payload into a stable-ordered
// slice. The API returns an object keyed by disruption ID (historically an
// empty array), so unknown shapes simply yield nil.
func (c Connection) DisruptionList() []Disruption {
	if len(c.Disruptions) == 0 {
		return nil
	}
	var byID map[string]Disruption
	if err := json.Unmarshal(c.Disruptions, &byID); err != nil {
		return nil
	}
	ids := make([]string, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	list := make([]Disruption, 0, len(ids))
	for _, id := range ids {
		list = append(list, byID[id])
	}
	return list
}

// FirstVehicleLeg returns the index of the first vehicle leg, or -1.
func (c Connection) FirstVehicleLeg() int {
	for i := range c.Legs {
		if c.Legs[i].IsVehicle() {
			return i
		}
	}
	return -1
}

// LastVehicleLeg returns the index of the last vehicle leg, or -1.
func (c Connection) LastVehicleLeg() int {
	for i := len(c.Legs) - 1; i >= 0; i-- {
		if c.Legs[i].IsVehicle() {
			return i
		}
	}
	return -1
}
