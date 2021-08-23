package validator

import (
	"encoding"
	"fmt"
	"strconv"

	"github.com/go-courier/httptransport/validator/rules"
	"github.com/go-courier/x/ptr"
)

func MinInt(bitSize uint) int64 {
	return -(1 << (bitSize - 1))
}

func MaxInt(bitSize uint) int64 {
	return 1<<(bitSize-1) - 1
}

func MaxUint(bitSize uint) uint64 {
	return 1<<bitSize - 1
}

func RangeFromUint(min uint64, max *uint64) []*rules.RuleLit {
	ranges := make([]*rules.RuleLit, 2)

	if min == 0 && max == nil {
		return nil
	}

	ranges[0] = rules.NewRuleLit([]byte(fmt.Sprintf("%d", min)))

	if max != nil {
		if min == *max {
			return []*rules.RuleLit{ranges[0]}
		}
		ranges[1] = rules.NewRuleLit([]byte(fmt.Sprintf("%d", *max)))
	}

	return ranges
}

func UintRange(typ string, bitSize uint, ranges ...*rules.RuleLit) (uint64, *uint64, error) {
	parseUint := func(b []byte) (*uint64, error) {
		if len(b) == 0 {
			return nil, nil
		}
		n, err := strconv.ParseUint(string(b), 10, int(bitSize))
		if err != nil {
			return nil, fmt.Errorf(" %s value is not correct: %s", typ, err)
		}
		return &n, nil
	}

	switch len(ranges) {
	case 2:
		min, err := parseUint(ranges[0].Bytes())
		if err != nil {
			return 0, nil, fmt.Errorf("min %s", err)
		}
		if min == nil {
			min = ptr.Uint64(0)
		}

		max, err := parseUint(ranges[1].Bytes())
		if err != nil {
			return 0, nil, fmt.Errorf("max %s", err)
		}

		if max != nil && *max < *min {
			return 0, nil, fmt.Errorf("max %s value must be equal or large than min value %d, current %d", typ, min, max)
		}

		return *min, max, nil
	case 1:
		min, err := parseUint(ranges[0].Bytes())
		if err != nil {
			return 0, nil, fmt.Errorf("min %s", err)
		}
		if min == nil {
			min = ptr.Uint64(0)
		}
		return *min, min, nil
	}
	return 0, nil, nil
}

func ToMarshalledText(v interface{}) string {
	if m, ok := v.(encoding.TextMarshaler); ok {
		d, _ := m.MarshalText()
		return string(d)
	}
	return ""
}
