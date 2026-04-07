package libgograbber

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestUnpackPortString(t *testing.T) {
	tests := []struct {
		input    string
		expected []int
	}{
		{"80", []int{80}},
		{"80,443", []int{80, 443}},
		{"80-82", []int{80, 81, 82}},
		{"80,82-84,443", []int{80, 82, 83, 84, 443}},
	}

	for _, tc := range tests {
		got := UnpackPortString(tc.input)
		for _, p := range tc.expected {
			if !got.Contains(p) {
				t.Errorf("UnpackPortString(%q) missing port %d", tc.input, p)
			}
		}
		if len(got.Set) != len(tc.expected) {
			t.Errorf("UnpackPortString(%q) got %d ports, expected %d", tc.input, len(got.Set), len(tc.expected))
		}
	}
}

func TestExpandHosts(t *testing.T) {
	tests := []struct {
		input    []string
		expected []string
	}{
		{[]string{"127.0.0.1"}, []string{"127.0.0.1"}},
		{[]string{"127.0.0.1/31"}, []string{"127.0.0.0", "127.0.0.1"}},
		{[]string{"example.com"}, []string{"example.com"}},
	}

	for _, tc := range tests {
		got := ExpandHosts(tc.input)
		for _, h := range tc.expected {
			if !got.Contains(h) {
				t.Errorf("ExpandHosts(%v) missing host %q", tc.input, h)
			}
		}
	}
}

func TestParseURLToHost(t *testing.T) {
	loggers := InitLogger(os.Stdout, os.Stdout, os.Stdout, os.Stdout, os.Stdout)
	tests := []struct {
		input    string
		expected Host
	}{
		{"http://example.com", Host{HostAddr: "example.com", Port: 80, Protocol: "http", Path: ""}},
		{"https://example.com:8443/test", Host{HostAddr: "example.com", Port: 8443, Protocol: "https", Path: "/test"}},
	}

	for _, tc := range tests {
		targets := make(chan Host, 1)
		ParseURLToHost(loggers, tc.input, targets)
		close(targets)
		got := <-targets
		if got.HostAddr != tc.expected.HostAddr || got.Port != tc.expected.Port || got.Protocol != tc.expected.Protocol || got.Path != tc.expected.Path {
			t.Errorf("ParseURLToHost(%q) = %+v, expected %+v", tc.input, got, tc.expected)
		}
	}
}

func TestStringSet(t *testing.T) {
	s := StringSet{Set: make(map[string]bool)}
	s.Add("test")
	if !s.Contains("test") {
		t.Error("StringSet should contain 'test'")
	}
	if s.Contains("missing") {
		t.Error("StringSet should not contain 'missing'")
	}
	s.AddRange([]string{"a", "b"})
	if !s.Contains("a") || !s.Contains("b") {
		t.Error("StringSet should contain 'a' and 'b'")
	}
	if !s.ContainsAny([]string{"b", "z"}) {
		t.Error("ContainsAny should return true")
	}
	if s.ContainsAny([]string{"x", "y"}) {
		t.Error("ContainsAny should return false")
	}
	if got := s.Stringify(); !strings.Contains(got, "a") || !strings.Contains(got, "b") || !strings.Contains(got, "test") {
		t.Errorf("Stringify got %q", got)
	}
}

func TestIntSet(t *testing.T) {
	s := IntSet{Set: make(map[int]bool)}
	s.Add(80)
	if !s.Contains(80) {
		t.Error("IntSet should contain 80")
	}
	if s.Contains(443) {
		t.Error("IntSet should not contain 443")
	}
	s.Add(443)
	if got := s.Stringify(); !strings.Contains(got, "80") || !strings.Contains(got, "443") {
		t.Errorf("IntSet.Stringify got %q", got)
	}
}

func TestChunkString(t *testing.T) {
	got := ChunkString("abcdef", 2)
	expected := []string{"ab", "cd", "ef"}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("ChunkString got %v, expected %v", got, expected)
	}
}

func TestPadding(t *testing.T) {
	if got := LeftPad2Len("test", " ", 10); got != "      test" {
		t.Errorf("LeftPad2Len got %q, expected %q", got, "      test")
	}
	if got := RightPad2Len("test", " ", 10); got != "test      " {
		t.Errorf("RightPad2Len got %q, expected %q", got, "test      ")
	}
}

func TestFileStuff(t *testing.T) {
	loggers := InitLogger(os.Stdout, os.Stdout, os.Stdout, os.Stdout, os.Stdout)
	content := "line1\nline2\nline3"
	tmpfile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	lines, err := readLines(tmpfile.Name())
	if err != nil {
		t.Errorf("readLines failed: %v", err)
	}
	expected := []string{"line1", "line2", "line3"}
	if !reflect.DeepEqual(lines, expected) {
		t.Errorf("readLines got %v, expected %v", lines, expected)
	}

	data, err := GetDataFromFile(loggers, tmpfile.Name())
	if err != nil {
		t.Errorf("GetDataFromFile failed: %v", err)
	}
	if !reflect.DeepEqual(data, expected) {
		t.Errorf("GetDataFromFile got %v, expected %v", data, expected)
	}
}

func TestRandomStuff(t *testing.T) {
	if got := StringWithCharset(10, "abc"); len(got) != 10 {
		t.Errorf("StringWithCharset length got %d, expected 10", len(got))
	}
	if got := RandString(); len(got) == 0 {
		t.Error("RandString returned empty string")
	}
}

func TestHostHashes(t *testing.T) {
	h := Host{HostAddr: "127.0.0.1", Port: 80, Protocol: "http"}
	hash := h.PrefetchHash()
	if len(hash) == 0 {
		t.Error("PrefetchHash returned empty")
	}
	hashes := map[string]bool{hash: true}
	if !h.PrefetchDoneCheck(hashes) {
		t.Error("PrefetchDoneCheck should be true")
	}

	s404hash := h.Soft404Hash()
	if len(s404hash) == 0 {
		t.Error("Soft404Hash returned empty")
	}
	s404hashes := map[string]bool{s404hash: true}
	if !h.Soft404DoneCheck(s404hashes) {
		t.Error("Soft404DoneCheck should be true")
	}
}
