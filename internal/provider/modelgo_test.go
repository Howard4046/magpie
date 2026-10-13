package provider

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// ModelGo's international and China sites are separate accounts and keys,
// each serving Chat, Responses and Messages at its own host. Its model list
// is the gateway's own, so the preset borrows no models.dev entry.
func TestModelGoPreset(t *testing.T) {
	pr := Preset("modelgo")
	if pr == nil || pr.Kind != KindRelay || pr.NoKey || pr.Sponsored || pr.Catalog != "" {
		t.Fatalf("a keyed relay preset, not sponsored, with no borrowed catalog: %+v", pr)
	}
	p, err := FromPreset("modelgo")
	if err != nil {
		t.Fatal(err)
	}
	if p.Chat != "https://api.modelgo.ai/v1" || p.Responses != p.Chat || p.Anthropic != "https://api.modelgo.ai" {
		t.Fatalf("international endpoints: %q %q %q", p.Chat, p.Responses, p.Anthropic)
	}
	// the preset's own endpoints are its first region's
	if len(pr.Regions) != 2 || pr.Regions[0].Chat != pr.Chat || pr.Regions[0].Anthropic != pr.Anthropic {
		t.Fatalf("regions: %+v", pr.Regions)
	}
	for _, r := range pr.Regions {
		host := map[string]string{"global": "api.modelgo.ai", "china": "api.modelgo.com"}[r.ID]
		if host == "" {
			t.Fatalf("region: %+v", r)
		}
		if r.Chat != "https://"+host+"/v1" || r.Responses != r.Chat || r.Anthropic != "https://"+host {
			t.Fatalf("region %s: %q %q %q", r.ID, r.Chat, r.Responses, r.Anthropic)
		}
	}
	// a key from one site is refused at the other, so each site's Get-a-key
	// link is its own console
	global := normalize(Provider{Preset: "modelgo", Chat: "https://api.modelgo.ai/v1", Anthropic: "https://api.modelgo.ai"})
	china := normalize(Provider{Preset: "modelgo", Chat: "https://api.modelgo.com/v1", Anthropic: "https://api.modelgo.com"})
	if global.KeysURL != "https://console.modelgo.ai" || china.KeysURL != "https://console.modelgo.com" {
		t.Fatalf("keys pages: %q, %q", global.KeysURL, china.KeysURL)
	}
	// the icon the preset names is a file the app ships
	if _, err := os.Stat(filepath.Join("..", "gui", "assets", "icons", pr.Icon+".svg")); err != nil {
		t.Fatalf("icon: %v", err)
	}
}

// A ModelGo key can be added from an import link, with the model slugs kept
// as the gateway names them, and an app already configured at one of ModelGo's
// documented bases is recognised as the preset.
func TestModelGoImport(t *testing.T) {
	detectHome(t)
	p, err := ParseImport("magpie://import?preset=modelgo&key=sk-test&models=claude-opus-5.5,gpt-5.6-sol")
	if err != nil {
		t.Fatal(err)
	}
	if id, err := Add(p); err != nil || id != "modelgo" {
		t.Fatalf("add: %q, %v", id, err)
	}
	got, err := Find("modelgo")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "ModelGo" || got.Icon != "modelgo-color" || got.Key != "sk-test" ||
		!slices.Equal(got.Models, []string{"claude-opus-5.5", "gpt-5.6-sol"}) {
		t.Fatalf("saved provider: %+v", got)
	}
	for _, c := range []struct {
		e    endpoints
		host string
	}{
		{endpoints{chat: "https://api.modelgo.ai/v1/"}, "https://api.modelgo.ai"},
		{endpoints{responses: "https://api.modelgo.ai/v1"}, "https://api.modelgo.ai"},
		{endpoints{anthropic: "https://api.modelgo.ai"}, "https://api.modelgo.ai"},
		{endpoints{chat: "https://api.modelgo.com/v1"}, "https://api.modelgo.com"},
		{endpoints{anthropic: "https://api.modelgo.com"}, "https://api.modelgo.com"},
	} {
		im, skip := imported("default", "sk-import", c.e, nil)
		if skip != "" || im.Preset != "modelgo" || im.Key != "sk-import" || im.Anthropic != c.host {
			t.Fatalf("import at %+v: %+v, %s", c.e, im, skip)
		}
	}
}
