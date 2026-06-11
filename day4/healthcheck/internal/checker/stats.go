package checker

import (
    "math"
    "sort"
    "time"
)

type Percentiles struct {
    P50  time.Duration
    P95  time.Duration
    P99  time.Duration
    Min  time.Duration
    Max  time.Duration
    Mean time.Duration
}

func Calculate(latencies []time.Duration) Percentiles {
    if len(latencies) == 0 {
        return Percentiles{}
    }
    sorted := make([]time.Duration, len(latencies))
    copy(sorted, latencies)
    sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

    var sum time.Duration
    for _, l := range sorted {
        sum += l
    }

    return Percentiles{
        P50:  percentile(sorted, 0.50),
        P95:  percentile(sorted, 0.95),
        P99:  percentile(sorted, 0.99),
        Min:  sorted[0],
        Max:  sorted[len(sorted)-1],
        Mean: sum / time.Duration(len(sorted)),
    }
}

func percentile(sorted []time.Duration, p float64) time.Duration {
    if len(sorted) == 0 {
        return 0
    }
    if p <= 0 {
        return sorted[0]
    }
    if p >= 1 {
        return sorted[len(sorted)-1]
    }
    idx := int(math.Ceil(p*float64(len(sorted)))) - 1
    if idx < 0 {
        idx = 0
    }
    return sorted[idx]
}
