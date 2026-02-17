package engine

import "encoding/json"

// WSMessage is the JSON message sent over WebSocket
type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// MarshalWSMessage creates a JSON-encoded WebSocket message
func MarshalWSMessage(msgType string, data interface{}) ([]byte, error) {
	msg := WSMessage{Type: msgType, Data: data}
	return json.Marshal(msg)
}

// Event data types

type StateChangeData struct {
	State string `json:"state"` // "idle", "running", "paused", "stopped"
}

type RollAttemptedData struct {
	AttemptNum  int     `json:"attemptNum"`
	MaxAttempts int     `json:"maxAttempts"`
	TotalRolls  int     `json:"totalRolls"`
	RollsPerMin float64 `json:"rollsPerMin"`
}

type TooltipCapturedData struct {
	Timestamp int64 `json:"timestamp"` // Unix ms
}

type ModsTrackedData struct {
	OCRText    string            `json:"ocrText"`
	ParsedMods map[string]int    `json:"parsedMods"` // mod name -> value
	ModStats   map[string]*ModStat `json:"modStats"`
	TotalRolls int               `json:"totalRolls"`
}

type TargetFoundData struct {
	ModName    string `json:"modName"`
	Value      int    `json:"value"`
	AttemptNum int    `json:"attemptNum"`
	TotalRolls int    `json:"totalRolls"`
}

type ItemStartedData struct {
	ItemNumber int `json:"itemNumber"`
	PendingX   int `json:"pendingX"`
	PendingY   int `json:"pendingY"`
}

type ItemCompletedData struct {
	ItemNumber int    `json:"itemNumber"`
	Success    bool   `json:"success"`
	ResultX    int    `json:"resultX"`
	ResultY    int    `json:"resultY"`
}

type CraftCountdownData struct {
	SecondsLeft int `json:"secondsLeft"`
}

type SnapshotUpdatedData struct {
	Filename   string `json:"filename"`
	StepName   string `json:"stepName"`
	ItemNumber int    `json:"itemNumber"`
}

type SessionEndedData struct {
	Report *ReportData `json:"report"`
}

type CaptureCountdownData struct {
	SecondsLeft int    `json:"secondsLeft"`
	Field       string `json:"field"`
}

type CaptureResultData struct {
	Field string `json:"field"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
}

// ReportData is a JSON-serializable version of session report
type ReportData struct {
	StartTime     string              `json:"startTime"`
	EndTime       string              `json:"endTime"`
	Duration      string              `json:"duration"`
	TotalRolls    int                 `json:"totalRolls"`
	RollsPerMin   float64             `json:"rollsPerMin"`
	TargetMods    []string            `json:"targetMods"`
	TargetModHit  bool                `json:"targetModHit"`
	TargetModName string              `json:"targetModName"`
	TargetValue   int                 `json:"targetValue"`
	ModStats      []ReportModStat     `json:"modStats"`
	RoundResults  []ReportRoundResult `json:"roundResults"`
}

type ReportModStat struct {
	ModName     string  `json:"modName"`
	Count       int     `json:"count"`
	MinValue    int     `json:"minValue"`
	MaxValue    int     `json:"maxValue"`
	AvgValue    float64 `json:"avgValue"`
	Probability float64 `json:"probability"` // percentage
}

type ReportRoundResult struct {
	RoundNumber   int    `json:"roundNumber"`
	Success       bool   `json:"success"`
	TargetHit     bool   `json:"targetHit"`
	TargetModName string `json:"targetModName"`
	TargetValue   int    `json:"targetValue"`
}
