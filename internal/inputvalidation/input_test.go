package inputvalidation

import "testing"

func TestIdentifierCompatibility(t *testing.T) {
	schema := []byte(`{"type":"object","properties":{"device_id":{"type":"string"}},"required":["device_id"]}`)
	for _, id := range []string{"507f1f77bcf86cd799439011", "f47ac10b-58cc-4372-a567-0e02b2c3d479", "legacy_123", "123", "default", "opaque:123", "resource.v2"} {
		if err := Validate(schema, []byte(`{"device_id":"`+id+`"}`)); err != nil {
			t.Errorf("%s: %v", id, err)
		}
	}
	for _, id := range []string{"", " ", ".", "..", "../foo", "foo/bar", "foo%2fbar", "foo?bar", "foo#bar", "a\\\\b"} {
		if err := Validate(schema, []byte(`{"device_id":"`+id+`"}`)); err == nil {
			t.Errorf("accepted %q", id)
		}
	}
}

func TestOnlyDeclaredRoutingStringsAreConstrained(t *testing.T) {
	schema := []byte(`{"type":"object","properties":{"port_id":{"type":"integer"},"group_id":{"type":"array","items":{"type":"integer"}},"mac":{"type":"string"},"config":{"type":"object"}}}`)
	input := []byte(`{"port_id":4,"group_id":[1,2],"mac":"","unknown_id":"opaque/value","config":{"network_id":"","mac":"any API-specific data","values":[1,true,null]}}`)
	if err := Validate(schema, input); err != nil {
		t.Fatal(err)
	}
	for _, mac := range []string{"AA:BB:CC:DD:EE:FF", "aa-bb-cc-dd-ee-ff", "aabb.ccdd.eeff"} {
		if err := Validate(schema, []byte(`{"mac":"`+mac+`"}`)); err == nil {
			t.Errorf("accepted unsupported MAC %s", mac)
		}
	}
}
