// 2.4 — 커스텀 Marshaler/Unmarshaler (전체 코드)
package main

import (
	"fmt"
	"strconv"
	"strings"
)

type Money int64 // cents 단위

func (m Money) MarshalJSON() ([]byte, error) {
	dollars := float64(m) / 100
	return []byte(fmt.Sprintf(`"%.2f"`, dollars)), nil
}

func (m *Money) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	*m = Money(f * 100)
	return nil
}
