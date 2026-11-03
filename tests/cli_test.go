package tests

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI(t *testing.T) {
	for _, tc := range []struct {
		name, q, status string
		extra           []string
	}{
		{"normal", "视频导出格式", "answered", nil},
		{"conflict", "报销期限多久", "conflict", nil},
		{"outside", "明天会下雨吗", "insufficient", nil},
		{"paraphrase", "费用票据最晚什么时候提交", "conflict", nil},
		{"followup", "那可以用哪些类型", "answered", []string{"-history", "用户正在咨询产品的视频导出"}},
		{"followup_without_history", "那可以用哪些类型", "insufficient", nil},
		{"failure", "报销期限", "tool_error", []string{"-fail", "read"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"run", "../src", "-docs", "../data/docs.json", "-q", tc.q}, tc.extra...)
			raw, err := exec.Command("go", args...).CombinedOutput()
			if err != nil {
				t.Fatalf("%v: %s", err, raw)
			}
			var r struct {
				Status, Answer string
				Citations      []struct{ ID, Quote string }
			}
			if err = json.Unmarshal(raw, &r); err != nil || r.Status != tc.status {
				t.Fatalf("%v: %s", err, raw)
			}
			for _, c := range r.Citations {
				if !strings.Contains(r.Answer, c.Quote) || c.ID == "" || c.Quote == "" {
					t.Fatalf("invalid evidence: %+v", c)
				}
			}
			if tc.status == "conflict" && len(r.Citations) != 2 {
				t.Fatal("expected both policies")
			}
		})
	}
}

func TestDocumentChanges(t *testing.T) {
	for _, tc := range []struct {
		second       bool
		days, status string
	}{
		{false, "30", "answered"}, {true, "30", "conflict"}, {false, "45", "answered"},
	} {
		docs := []map[string]string{{"id": "D01", "date": "2026-01-01", "tags": "报销", "quote": "报销期限为 " + tc.days + " 天。"}}
		if tc.second {
			docs = append(docs, map[string]string{"id": "D02", "date": "2026-02-01", "tags": "报销", "quote": "报销期限为 15 天。"})
		}
		raw, err := json.Marshal(docs)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(t.TempDir(), "docs.json")
		if err = os.WriteFile(path, raw, 0600); err != nil {
			t.Fatal(err)
		}
		out, err := exec.Command("go", "run", "../src", "-docs", path, "-q", "报销期限").CombinedOutput()
		if err != nil {
			t.Fatalf("%v: %s", err, out)
		}
		var result struct {
			Status, Answer string
			Citations      []struct{ ID, Quote string }
			Trace          []struct{ Action struct{ Tool, Arg string } }
		}
		if err = json.Unmarshal(out, &result); err != nil {
			t.Fatal(err)
		}
		if result.Status != tc.status || !strings.Contains(result.Answer, docs[0]["quote"]) || len(result.Citations) != len(docs) {
			t.Fatalf("%s", out)
		}
		checked := false
		for _, step := range result.Trace {
			if step.Action.Tool == "search" && step.Action.Arg == "适用范围 优先级" {
				checked = true
			}
		}
		if checked != tc.second {
			t.Fatalf("unexpected follow-up search: %s", out)
		}
		for i, d := range docs {
			if result.Citations[i].ID != d["id"] || result.Citations[i].Quote != d["quote"] {
				t.Fatalf("%s", out)
			}
		}
	}
}
