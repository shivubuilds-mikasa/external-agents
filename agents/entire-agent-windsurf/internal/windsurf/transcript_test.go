package windsurf

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseWindsurfTranscriptFixture(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "cascade_transcript.jsonl"))
	if err != nil {
		t.Fatal(err)
	}

	transcript, err := parseWindsurfTranscript(data)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"create a hello world file"}; !reflect.DeepEqual(transcript.prompts, want) {
		t.Fatalf("prompts = %#v, want %#v", transcript.prompts, want)
	}
	if want := []string{"I'll create a hello world file for you.", "I created the file for you."}; !reflect.DeepEqual(transcript.responses, want) {
		t.Fatalf("responses = %#v, want %#v", transcript.responses, want)
	}
	if want := []string{"/path/to/file.py"}; !reflect.DeepEqual(transcript.modifiedFiles, want) {
		t.Fatalf("modified files = %#v, want %#v", transcript.modifiedFiles, want)
	}
}

func TestParseWindsurfTranscriptToleratesIncompleteRecords(t *testing.T) {
	data := []byte("\n" +
		`{"type":"user_input","user_input":{"user_response":"first prompt"}}` + "\n" +
		`{"type":"planner_response","planner_response":{"response":"first response"}}` + "\n" +
		`{"type":"code_action","code_action":{"path":" /repo/a.go "}}` + "\n" +
		`{"type":"code_action","code_action":{"path":"/repo/a.go"}}` + "\n" +
		`{"type":"code_action","code_action":{"path":"/repo/b.go"}}` + "\n" +
		`{"type":"unknown","unknown":{"value":true}}` + "\n" +
		`{"type":"user_input","user_input":{}}` + "\n" +
		`{"type":"planner_response","planner_response":{}}` + "\n" +
		`{"type":"code_action","code_action":{}}` + "\n" +
		`{"type":` + "\n" +
		`{"type":"user_input","user_input":{"user_response":"second prompt"}}` + "\n" +
		`{"type":"planner_response","planner_response":{"response":"second response"}}` + "\n")

	transcript, err := parseWindsurfTranscript(data)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"first prompt", "second prompt"}; !reflect.DeepEqual(transcript.prompts, want) {
		t.Fatalf("prompts = %#v, want %#v", transcript.prompts, want)
	}
	if want := []string{"first response", "second response"}; !reflect.DeepEqual(transcript.responses, want) {
		t.Fatalf("responses = %#v, want %#v", transcript.responses, want)
	}
	if want := []string{"/repo/a.go", "/repo/b.go"}; !reflect.DeepEqual(transcript.modifiedFiles, want) {
		t.Fatalf("modified files = %#v, want %#v", transcript.modifiedFiles, want)
	}
}

func TestParseWindsurfTranscriptEmpty(t *testing.T) {
	transcript, err := parseWindsurfTranscript([]byte(" \n\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(transcript.prompts) != 0 || len(transcript.responses) != 0 || len(transcript.modifiedFiles) != 0 {
		t.Fatalf("transcript = %#v", transcript)
	}
}
