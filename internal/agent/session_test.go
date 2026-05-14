package agent

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/electr1fy0/sorta/internal/llm"
)

type stubClient struct {
	responses []string
	err       error
	index     int
}

func (c *stubClient) Run(ctx context.Context, req llm.Request) (string, error) {
	if c.err != nil {
		return "", c.err
	}
	if c.index >= len(c.responses) {
		return "", nil
	}
	resp := c.responses[c.index]
	c.index++
	return resp, nil
}

func TestSessionPlanParsesJSONPlan(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "os_notes.pdf"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	client := &stubClient{
		responses: []string{`{"summary":"Group OS material and create pyqs","actions":[{"kind":"mkdir","mkdir":{"path":"pyqs"}},{"kind":"sort_rule","sort_rule":{"folder":"OS","keywords":["os","operating systems"]}}]}`},
	}

	session := NewSession(client, SessionOptions{})
	plan, err := session.Plan(context.Background(), dir, "group files", "model", nil)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}

	if len(plan.Actions) != 2 {
		t.Fatalf("expected 2 planned actions, got %d", len(plan.Actions))
	}
	if plan.Actions[1].Kind != ActionSortRule {
		t.Fatalf("expected second action sort rule, got %s", plan.Actions[1].Kind)
	}
}

func TestSessionPlanFailsOnInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	session := NewSession(&stubClient{responses: []string{`not json`}}, SessionOptions{})
	_, err := session.Plan(context.Background(), dir, "test", "model", nil)
	if err == nil || !strings.Contains(err.Error(), "parse plan JSON") {
		t.Fatalf("expected parse error, got %v", err)
	}
}

func TestSessionPlanFailsOnInvalidAction(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	session := NewSession(&stubClient{responses: []string{`{"summary":"x","actions":[{"kind":"mkdir","mkdir":{"path":"../oops"}}]}`}}, SessionOptions{})
	_, err := session.Plan(context.Background(), dir, "test", "model", nil)
	if err == nil || !strings.Contains(err.Error(), "invalid mkdir path") {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestSessionPlanCanRequestFilenames(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	session := NewSession(&stubClient{responses: []string{`{"summary":"Need examples to infer sorting rules","request_filenames":["Give 3 example filenames"],"actions":[]}`}}, SessionOptions{})
	plan, err := session.Plan(context.Background(), dir, "sort by subject", "model", nil)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if len(plan.RequestFilenames) != 1 || plan.RequestFilenames[0] != "Give 3 example filenames" {
		t.Fatalf("unexpected filename requests: %#v", plan.RequestFilenames)
	}
}
