package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRemoteLoop(t *testing.T) {
	actions := []Action{
		{Tool: "search", Arg: "报销"},
		{Tool: "read", Arg: "D01"},
		{Tool: "finish", Status: "answered", Answer: "30 天", IDs: []string{"D01"}},
	}
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Messages []struct{ Content string } }
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Messages) != 2 || calls >= len(actions) {
			http.Error(w, "bad request", 400)
			return
		}
		for i, expected := range []string{"报销期限多久", "D01", "报销期限为 30 天。"} {
			if i <= calls && !strings.Contains(req.Messages[1].Content, expected) {
				t.Errorf("step %d missing %q", calls, expected)
			}
		}
		content, _ := json.Marshal(actions[calls])
		calls++
		json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"content": string(content)}}}})
	}))
	defer server.Close()
	t.Setenv("AGENT_ENDPOINT", server.URL)
	t.Setenv("AGENT_MODEL", "test")
	t.Setenv("AGENT_API_KEY", "")
	result := run(context.Background(), []Doc{{ID: "D01", Date: "2026-01-01", Tags: "报销", Quote: "报销期限为 30 天。"}}, "报销期限多久", "", "", remote)
	if result.Status != "answered" || len(result.Citations) != 1 || calls != 3 {
		t.Fatalf("%+v", result)
	}
}

func TestExecutionGuards(t *testing.T) {
	for _, tc := range []struct {
		action Action
		gap    string
		steps  int
	}{
		{Action{Tool: "finish", Status: "answered", IDs: []string{"D99"}}, "引用必须来自本轮读取的不同文档", 1},
		{Action{Tool: "search", Arg: "未知"}, "达到决策步数上限", 12},
	} {
		result := run(context.Background(), nil, "问题", "", "", func(context.Context, string, string, []Step) (Action, error) { return tc.action, nil })
		if result.Status != "incomplete" || result.Gap != tc.gap || len(result.Trace) != tc.steps || len(result.Citations) != 0 {
			t.Fatalf("%+v", result)
		}
	}
}

func TestRemoteFailures(t *testing.T) {
	for _, tc := range []struct {
		name    string
		gap     string
		handler http.HandlerFunc
		timeout time.Duration
	}{
		{"invalid_action", "模型动作 JSON 无效", func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"invalid"}}]}`))
		}, time.Second},
		{"timeout", "模型请求失败或超时", func(_ http.ResponseWriter, r *http.Request) {
			select {
			case <-r.Context().Done():
			case <-time.After(time.Second):
			}
		}, 50 * time.Millisecond},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			t.Setenv("AGENT_ENDPOINT", server.URL)
			t.Setenv("AGENT_MODEL", "test")
			t.Setenv("AGENT_API_KEY", "")
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()
			result := run(ctx, nil, "报销期限多久", "", "", remote)
			if result.Status != "incomplete" || !strings.Contains(result.Gap, tc.gap) || len(result.Citations) != 0 {
				t.Fatalf("unexpected result: %+v", result)
			}
		})
	}
}
