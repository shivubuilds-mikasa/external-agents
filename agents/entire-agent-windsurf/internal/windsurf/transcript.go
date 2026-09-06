package windsurf

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

const maxTranscriptLine = 10 * 1024 * 1024

// parsedTranscript contains the documented Cascade transcript information
// needed by the adapter. It is intentionally internal: the protocol continues
// to receive the original JSONL through AgentSessionJSON.NativeData.
type parsedTranscript struct {
	prompts       []string
	responses     []string
	modifiedFiles []string
}

type cascadeTranscriptRecord struct {
	Type            string `json:"type"`
	UserInput       struct {
		UserResponse string `json:"user_response"`
	} `json:"user_input"`
	PlannerResponse struct {
		Response string `json:"response"`
	} `json:"planner_response"`
	CodeAction struct {
		Path string `json:"path"`
	} `json:"code_action"`
}

// parseWindsurfTranscript tolerates incomplete or partially corrupt JSONL,
// matching the other adapters' treatment of individual malformed records.
// Scanner errors, such as a record that exceeds the bounded line size, remain
// actionable because the stream cannot be safely continued in that case.
func parseWindsurfTranscript(data []byte) (parsedTranscript, error) {
	result := parsedTranscript{modifiedFiles: []string{}}
	seenFiles := make(map[string]bool)

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 64*1024), maxTranscriptLine+1)
	line := 0
	for scanner.Scan() {
		line++
		if len(scanner.Bytes()) > maxTranscriptLine {
			return parsedTranscript{}, fmt.Errorf("scan Windsurf transcript: line %d exceeds %d bytes", line, maxTranscriptLine)
		}

		var record cascadeTranscriptRecord
		if json.Unmarshal(scanner.Bytes(), &record) != nil {
			continue
		}

		switch record.Type {
		case "user_input":
			if prompt := strings.TrimSpace(record.UserInput.UserResponse); prompt != "" {
				result.prompts = append(result.prompts, prompt)
			}
		case "planner_response":
			if response := strings.TrimSpace(record.PlannerResponse.Response); response != "" {
				result.responses = append(result.responses, response)
			}
		case "code_action":
			if path := strings.TrimSpace(record.CodeAction.Path); path != "" && !seenFiles[path] {
				seenFiles[path] = true
				result.modifiedFiles = append(result.modifiedFiles, path)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return parsedTranscript{}, fmt.Errorf("scan Windsurf transcript: %w", err)
	}
	return result, nil
}
