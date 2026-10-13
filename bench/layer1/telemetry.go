package layer1

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"sort"
)

// TelemetrySchemaID names the record format.
const TelemetrySchemaID = "gira-bench/layer1/telemetry/1"

// Telemetry is one run record. Every field listed in spec B, Layer 1, is present.
type Telemetry struct {
	Schema          string         `json:"schema"`
	TaskID          string         `json:"task_id"`
	Arm             string         `json:"arm"`
	ModelID         string         `json:"model_id"`
	Repetition      int            `json:"repetition"`
	CorpusSHA       string         `json:"corpus_sha"`
	Temperature     float64        `json:"temperature"`
	TokensInput     int            `json:"tokens_input"`
	TokensOutput    int            `json:"tokens_output"`
	CacheReadTokens int            `json:"cache_read_tokens"`
	ToolCalls       int            `json:"tool_calls"`
	ToolCallsByTool map[string]int `json:"tool_calls_by_tool"`
	Turns           int            `json:"turns"`
	WallLatencyMS   int64          `json:"wall_latency_ms"`
	Precision       float64        `json:"precision"`
	Recall          float64        `json:"recall"`
	F1              float64        `json:"f1"`
	LocatorValidity float64        `json:"locator_validity"`
	LocatorsTotal   int            `json:"locators_total"`
	Hallucinated    []string       `json:"hallucinated_nodes"`
	Answered        bool           `json:"answered"`
	Error           string         `json:"error"`
}

//go:embed telemetry.schema.json
var telemetrySchema []byte

// ValidateTelemetry checks a JSON record against telemetry.schema.json: the
// object's required properties, the declared property types and
// additionalProperties false. This is the subset of JSON Schema the file uses.
func ValidateTelemetry(record []byte) error {
	var schema struct {
		Required             []string                         `json:"required"`
		Properties           map[string]struct{ Type string } `json:"properties"`
		AdditionalProperties *bool                            `json:"additionalProperties"`
	}
	if err := json.Unmarshal(telemetrySchema, &schema); err != nil {
		return err
	}
	var obj map[string]any
	if err := json.Unmarshal(record, &obj); err != nil {
		return fmt.Errorf("record is not a JSON object: %w", err)
	}
	for _, k := range schema.Required {
		if _, ok := obj[k]; !ok {
			return fmt.Errorf("missing required field %q", k)
		}
	}
	for k, v := range obj {
		p, ok := schema.Properties[k]
		if !ok {
			if schema.AdditionalProperties != nil && !*schema.AdditionalProperties {
				return fmt.Errorf("unexpected field %q", k)
			}
			continue
		}
		if !typeOK(p.Type, v) {
			return fmt.Errorf("field %q is not of type %s", k, p.Type)
		}
	}
	return nil
}

func typeOK(t string, v any) bool {
	switch t {
	case "string":
		_, ok := v.(string)
		return ok
	case "number":
		_, ok := v.(float64)
		return ok
	case "integer":
		f, ok := v.(float64)
		return ok && f == math.Trunc(f)
	case "boolean":
		_, ok := v.(bool)
		return ok
	case "object":
		_, ok := v.(map[string]any)
		return ok
	case "array":
		_, ok := v.([]any)
		return ok
	}
	return true
}

// Summary is the median and interquartile range of one metric over a cell's runs.
type Summary struct {
	Median float64 `json:"median"`
	IQR    float64 `json:"iqr"`
	N      int     `json:"n"`
}

// Summarize computes the median and IQR (linear interpolation between ranks).
func Summarize(xs []float64) Summary {
	if len(xs) == 0 {
		return Summary{}
	}
	s := append([]float64(nil), xs...)
	sort.Float64s(s)
	q := func(p float64) float64 {
		pos := p * float64(len(s)-1)
		lo := int(math.Floor(pos))
		hi := int(math.Ceil(pos))
		return s[lo] + (s[hi]-s[lo])*(pos-float64(lo))
	}
	return Summary{Median: q(0.5), IQR: q(0.75) - q(0.25), N: len(s)}
}
