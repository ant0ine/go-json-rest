package trie

import (
	"testing"
)

func TestPathInsert(t *testing.T) {

	trie := New()
	if trie.root == nil {
		t.Error()
	}

	trie.AddRoute("GET", "/", "1")
	if trie.root.Children["/"] == nil {
		t.Error()
	}

	trie.AddRoute("GET", "/r", "2")
	if trie.root.Children["/"].Children["r"] == nil {
		t.Error()
	}

	trie.AddRoute("GET", "/r/", "3")
	if trie.root.Children["/"].Children["r"].Children["/"] == nil {
		t.Error()
	}
}

func TestTrieCompression(t *testing.T) {

	trie := New()
	trie.AddRoute("GET", "/abc", "3")
	trie.AddRoute("GET", "/adc", "3")

	// before compression
	if trie.root.Children["/"].Children["a"].Children["b"].Children["c"] == nil {
		t.Error()
	}
	if trie.root.Children["/"].Children["a"].Children["d"].Children["c"] == nil {
		t.Error()
	}

	trie.Compress()

	// after compression
	if trie.root.Children["/abc"] == nil {
		t.Errorf("%+v", trie.root)
	}
	if trie.root.Children["/adc"] == nil {
		t.Errorf("%+v", trie.root)
	}

}
func TestParamInsert(t *testing.T) {
	trie := New()

	trie.AddRoute("GET", "/:id/", "")
	if trie.root.Children["/"].ParamChild.Children["/"] == nil {
		t.Error()
	}
	if trie.root.Children["/"].ParamName != "id" {
		t.Error()
	}

	trie.AddRoute("GET", "/:id/:property.:format", "")
	if trie.root.Children["/"].ParamChild.Children["/"].ParamChild.Children["."].ParamChild == nil {
		t.Error()
	}
	if trie.root.Children["/"].ParamName != "id" {
		t.Error()
	}
	if trie.root.Children["/"].ParamChild.Children["/"].ParamName != "property" {
		t.Error()
	}
	if trie.root.Children["/"].ParamChild.Children["/"].ParamChild.Children["."].ParamName != "format" {
		t.Error()
	}
}

func TestSplatInsert(t *testing.T) {
	trie := New()
	trie.AddRoute("GET", "/*splat", "")
	if trie.root.Children["/"].SplatChild == nil {
		t.Error()
	}
}

func TestDupeInsert(t *testing.T) {
	trie := New()
	trie.AddRoute("GET", "/", "1")
	err := trie.AddRoute("GET", "/", "2")
	if err == nil {
		t.Error()
	}
	if trie.root.Children["/"].HttpMethodToRoute["GET"] != "1" {
		t.Error()
	}
}

func isInMatches(test string, matches []*Match) bool {
	for _, match := range matches {
		if match.Route.(string) == test {
			return true
		}
	}
	return false
}

func TestFindRoute(t *testing.T) {

	trie := New()

	trie.AddRoute("GET", "/", "root")
	trie.AddRoute("GET", "/r/:id", "resource")
	trie.AddRoute("GET", "/r/:id/property", "property")
	trie.AddRoute("GET", "/r/:id/property.*format", "property_format")

	trie.Compress()

	matches := trie.FindRoutes("GET", "/")
	if len(matches) != 1 {
		t.Errorf("expected one route, got %d", len(matches))
	}
	if !isInMatches("root", matches) {
		t.Error("expected 'root'")
	}

	matches = trie.FindRoutes("GET", "/notfound")
	if len(matches) != 0 {
		t.Errorf("expected zero route, got %d", len(matches))
	}

	matches = trie.FindRoutes("GET", "/r/1")
	if len(matches) != 1 {
		t.Errorf("expected one route, got %d", len(matches))
	}
	if !isInMatches("resource", matches) {
		t.Errorf("expected 'resource', got %+v", matches)
	}
	if matches[0].Params["id"] != "1" {
		t.Error()
	}

	matches = trie.FindRoutes("GET", "/r/1/property")
	if len(matches) != 1 {
		t.Errorf("expected one route, got %d", len(matches))
	}
	if !isInMatches("property", matches) {
		t.Error("expected 'property'")
	}
	if matches[0].Params["id"] != "1" {
		t.Error()
	}

	matches = trie.FindRoutes("GET", "/r/1/property.json")
	if len(matches) != 1 {
		t.Errorf("expected one route, got %d", len(matches))
	}
	if !isInMatches("property_format", matches) {
		t.Error("expected 'property_format'")
	}
	if matches[0].Params["id"] != "1" {
		t.Error()
	}
	if matches[0].Params["format"] != "json" {
		t.Error()
	}
}

func TestFindRouteMultipleMatches(t *testing.T) {

	trie := New()

	trie.AddRoute("GET", "/r/1", "resource1")
	trie.AddRoute("GET", "/r/2", "resource2")
	trie.AddRoute("GET", "/r/:id", "resource_generic")
	trie.AddRoute("GET", "/s/*rest", "special_all")
	trie.AddRoute("GET", "/s/:param", "special_generic")
	trie.AddRoute("GET", "/", "root")

	trie.Compress()

	matches := trie.FindRoutes("GET", "/r/1")
	if len(matches) != 2 {
		t.Errorf("expected two matches, got %d", len(matches))
	}
	if !isInMatches("resource_generic", matches) {
		t.Error()
	}
	if !isInMatches("resource1", matches) {
		t.Error()
	}

	matches = trie.FindRoutes("GET", "/s/1")
	if len(matches) != 2 {
		t.Errorf("expected two matches, got %d", len(matches))
	}
	if !isInMatches("special_all", matches) {
		t.Error()
	}
	if !isInMatches("special_generic", matches) {
		t.Error()
	}
}

func TestConsistentPlaceholderName(t *testing.T) {

	trie := New()

	trie.AddRoute("GET", "/r/:id", "oneph")
	defer func() {
		if r := recover(); r == nil {
			t.Error("Should have died on adding second route")
		}
	}()
	trie.AddRoute("GET", "/r/:rid/other", "twoph")
}

func TestDuplicateName(t *testing.T) {

	trie := New()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Should have died, this route has two `:id`")
		}
	}()
	trie.AddRoute("GET", "/r/:id/o/:id", "two")
}
