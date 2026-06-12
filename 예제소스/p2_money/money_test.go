package main

import (
	"encoding/json"
	"testing"
)

func TestMoneyRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		cents Money
		want  string
	}{
		{"정수 달러", 1200, `"12.00"`},
		{"센트 포함", 1234, `"12.34"`},
		{"0원", 0, `"0.00"`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.cents)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != tc.want {
				t.Errorf("Marshal = %s, want %s", data, tc.want)
			}
			// 왕복(round-trip) 검증
			var back Money
			if err := json.Unmarshal(data, &back); err != nil {
				t.Fatal(err)
			}
			if back != tc.cents {
				t.Errorf("round-trip = %d, want %d", back, tc.cents)
			}
		})
	}
}

func TestMoneyUnmarshalInvalid(t *testing.T) {
	var m Money
	if err := json.Unmarshal([]byte(`"abc"`), &m); err == nil {
		t.Error("잘못된 입력인데 에러가 없음")
	}
}
