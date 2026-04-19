package structuredrewrite

import "testing"

func TestParseStructuredRewriteResponse(t *testing.T) {
	result, err := Parse(`{
		"rewrittenText": "Tin tiếng Việt",
		"rewrittenTextEn": "English news",
		"tweetVI": "Tweet VI",
		"tweetEN": "Tweet EN",
		"factCheckNote": "Đã kiểm chứng",
		"shouldPublish": true
	}`)
	if err != nil {
		t.Fatalf("parse structured response: %v", err)
	}
	if result.RewrittenText != "Tin tiếng Việt" {
		t.Fatalf("unexpected rewritten text: %s", result.RewrittenText)
	}
	if result.TweetTextVI != "Tweet VI" {
		t.Fatalf("unexpected tweet vi: %s", result.TweetTextVI)
	}
	if !result.ShouldPublish {
		t.Fatal("expected should publish")
	}
}

func TestParseStructuredRewriteResponseWithEmbeddedJSON(t *testing.T) {
	result, err := Parse("```json\n{\"rewrittenText\":\"Tin\",\"shouldPublish\":true}\n```")
	if err != nil {
		t.Fatalf("parse embedded json: %v", err)
	}
	if result.RewrittenText != "Tin" {
		t.Fatalf("unexpected rewritten text: %s", result.RewrittenText)
	}
}

func TestParseStructuredRewriteResponseAllowsEditorialSkip(t *testing.T) {
	result, err := Parse(`{"shouldPublish":false,"reason":"spam"}`)
	if err != nil {
		t.Fatalf("parse skip response: %v", err)
	}
	if result.ShouldPublish {
		t.Fatal("expected should publish false")
	}
	if result.Reason != "spam" {
		t.Fatalf("unexpected reason: %s", result.Reason)
	}
}
