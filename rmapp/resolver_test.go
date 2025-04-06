package rmapp

import "testing"

func TestGetBundle(t *testing.T) {
	expected := "org.wireshark.Wireshark"
	actual := extractQuotedSubstring("kMDItemCFBundleIdentifier = \"org.wireshark.Wireshark\"")
	if actual != expected {
		t.Errorf("Expected %s but got %s", expected, actual)
	}
}
