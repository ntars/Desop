package app

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseCommand(t *testing.T) {
	cases := []struct {
		args   []string
		action string
		comps  []string
		bad    bool
	}{
		{nil, "start", nil, false},
		{[]string{"status"}, "status", nil, false},
		{[]string{"selftest"}, "selftest", nil, false},
		{[]string{"selftest", "3"}, "selftest", nil, false},
		{[]string{"enable", "dock, widgets"}, "enable", []string{"dock", "widgets"}, false},
		{[]string{"disable", "island"}, "disable", []string{"island"}, false},
		{[]string{"BOGUS"}, "", nil, true},
		{[]string{"enable"}, "", nil, true},
		{[]string{"enable", "dock,dock"}, "", nil, true},
		{[]string{"status", "extra"}, "", nil, true},
		{[]string{"selftest", "0"}, "", nil, true},
	}
	for i, tc := range cases {
		got, err := ParseCommand(tc.args)
		if tc.bad {
			if err == nil {
				t.Errorf("case %d: expected error", i)
			}
			continue
		}
		if err != nil || got.Action != tc.action || !reflect.DeepEqual(got.Components, tc.comps) {
			t.Errorf("case %d: got %+v, err %v", i, got, err)
		}
	}
}

func TestStopEventNameStableAndPathSpecific(t *testing.T) {
	a := StopEventName(`C:\Users\Alice\AppData\Roaming\Desop\config.json`)
	b := StopEventName(`c:\users\alice\appdata\roaming\desop\config.json`)
	c := StopEventName(`C:\Users\Bob\AppData\Roaming\Desop\config.json`)
	if a != b {
		t.Fatalf("event name should be case-insensitive: %s vs %s", a, b)
	}
	if a == c {
		t.Fatal("different users must have different event names")
	}
	if !strings.HasPrefix(a, `Local\Desop_Stop_`) {
		t.Fatalf("unexpected event name: %s", a)
	}
}
