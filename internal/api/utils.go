package api

import (
	"encoding/json"
	"errors"
	"math"
	"strconv"
)

func isFinite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }

func parseJSONNumber(n json.Number) (float64, error) {
	f, err := n.Float64()
	if err != nil && !errors.Is(err, strconv.ErrRange) {
		return 0, err
	}
	return f, nil
}
