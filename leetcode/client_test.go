package leetcode_test

import (
	"strings"
	"testing"

	"github.com/conallob/coding-interview-pop-quiz/leetcode"
)

func TestHTMLToTextEmpty(t *testing.T) {
	if got := leetcode.HTMLToText(""); got != "" {
		t.Errorf("HTMLToText(\"\") = %q, want empty string", got)
	}
}

func TestHTMLToTextPlainText(t *testing.T) {
	got := leetcode.HTMLToText("hello world")
	if got != "hello world" {
		t.Errorf("HTMLToText(plain) = %q, want %q", got, "hello world")
	}
}

func TestHTMLToTextParagraphs(t *testing.T) {
	input := "<p>First paragraph.</p><p>Second paragraph.</p>"
	got := leetcode.HTMLToText(input)
	if !strings.Contains(got, "First paragraph.") {
		t.Errorf("missing first paragraph in %q", got)
	}
	if !strings.Contains(got, "Second paragraph.") {
		t.Errorf("missing second paragraph in %q", got)
	}
}

func TestHTMLToTextBreakTag(t *testing.T) {
	input := "line one<br/>line two"
	got := leetcode.HTMLToText(input)
	if !strings.Contains(got, "\n") {
		t.Errorf("expected newline from <br/>, got %q", got)
	}
	if !strings.Contains(got, "line one") || !strings.Contains(got, "line two") {
		t.Errorf("missing content: %q", got)
	}
}

func TestHTMLToTextEntitiesUnescaped(t *testing.T) {
	input := "<p>Use &lt;p&gt; tags and the &amp; symbol. &quot;Quoted&quot;.</p>"
	got := leetcode.HTMLToText(input)
	if !strings.Contains(got, "<p>") {
		t.Errorf("&lt;p&gt; not unescaped in %q", got)
	}
	if !strings.Contains(got, "&") {
		t.Errorf("&amp; not unescaped in %q", got)
	}
	if !strings.Contains(got, `"Quoted"`) {
		t.Errorf("&quot; not unescaped in %q", got)
	}
}

func TestHTMLToTextPreBlock(t *testing.T) {
	input := "<p>Example:</p><pre>Input: [1,2,3]\nOutput: 6</pre>"
	got := leetcode.HTMLToText(input)
	if !strings.Contains(got, "```") {
		t.Errorf("pre block should be wrapped in backtick fences, got %q", got)
	}
	if !strings.Contains(got, "Input: [1,2,3]") {
		t.Errorf("pre block content missing from %q", got)
	}
	if !strings.Contains(got, "Output: 6") {
		t.Errorf("pre block content missing from %q", got)
	}
}

func TestHTMLToTextPreWithInnerCode(t *testing.T) {
	input := "<pre><code>nums = [1, 2, 3]</code></pre>"
	got := leetcode.HTMLToText(input)
	if !strings.Contains(got, "```") {
		t.Errorf("pre/code block should use backtick fence, got %q", got)
	}
	if !strings.Contains(got, "nums = [1, 2, 3]") {
		t.Errorf("code content missing from %q", got)
	}
}

func TestHTMLToTextListItems(t *testing.T) {
	input := "<ul><li>First</li><li>Second</li><li>Third</li></ul>"
	got := leetcode.HTMLToText(input)
	if !strings.Contains(got, "•") {
		t.Errorf("list items should use bullet •, got %q", got)
	}
	if !strings.Contains(got, "First") || !strings.Contains(got, "Second") || !strings.Contains(got, "Third") {
		t.Errorf("list content missing from %q", got)
	}
}

func TestHTMLToTextStripsAllTags(t *testing.T) {
	inputs := []string{
		"<strong>bold</strong>",
		"<em>italic</em>",
		"<span class=\"foo\">text</span>",
		"<div>content</div>",
		"<h2>heading</h2>",
	}
	for _, input := range inputs {
		got := leetcode.HTMLToText(input)
		if strings.Contains(got, "<") || strings.Contains(got, ">") {
			t.Errorf("HTMLToText(%q) still contains angle brackets: %q", input, got)
		}
	}
}

func TestHTMLToTextCollapsesExcessiveNewlines(t *testing.T) {
	input := "<p>A</p><p></p><p></p><p>B</p>"
	got := leetcode.HTMLToText(input)
	if strings.Contains(got, "\n\n\n") {
		t.Errorf("expected collapsed newlines, got %q", got)
	}
}

func TestHTMLToTextTrimmed(t *testing.T) {
	input := "   <p>hello</p>   "
	got := leetcode.HTMLToText(input)
	if got != strings.TrimSpace(got) {
		t.Errorf("result not trimmed: %q", got)
	}
}

func TestHTMLToTextPreservesCodeContent(t *testing.T) {
	// Verify that HTML entities inside <pre> blocks are also unescaped.
	input := "<pre>if x &lt; y:</pre>"
	got := leetcode.HTMLToText(input)
	if !strings.Contains(got, "if x < y:") {
		t.Errorf("entities inside pre block not unescaped: %q", got)
	}
}
