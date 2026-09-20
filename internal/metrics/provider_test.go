package metrics

import (
	"testing"

	"github.com/statusengine/interface/internal/domain"
)

func TestBucketForKeepsWithinBudget(t *testing.T) {
	const budget = 500

	cases := map[string]int64{
		"an hour":   3600,
		"six hours": 6 * 3600,
		"a day":     86400,
		"a week":    7 * 86400,
		"a month":   30 * 86400,
		"a year":    365 * 86400,
	}
	for name, span := range cases {
		t.Run(name, func(t *testing.T) {
			bucket := BucketFor(0, span, budget)
			if bucket <= 0 {
				t.Fatalf("bucket = %d", bucket)
			}
			if points := span / bucket; points > budget {
				t.Errorf("%d points for a %ds window at %ds buckets, budget is %d",
					points, span, bucket, budget)
			}
		})
	}
}

// Picking a bigger bucket than needed throws away detail the budget
// could have paid for.
func TestBucketForPicksTheSmallestThatFits(t *testing.T) {
	const budget = 500
	span := int64(6 * 3600)

	bucket := BucketFor(0, span, budget)
	for _, smaller := range Buckets {
		if smaller >= bucket {
			break
		}
		if span/smaller <= budget {
			t.Errorf("chose %ds but %ds would also have fit", bucket, smaller)
		}
	}
}

func TestBucketForRoundNumbers(t *testing.T) {
	// "3m 17s average" is a label nobody can reason about.
	bucket := BucketFor(0, 12*3600, 500)
	for _, known := range Buckets {
		if bucket == known {
			return
		}
	}
	t.Errorf("bucket %ds is not one of the round resolutions", bucket)
}

func TestBucketForEdgeCases(t *testing.T) {
	if got := BucketFor(0, 0, 500); got != Buckets[0] {
		t.Errorf("zero window gave %d, want the finest bucket", got)
	}
	if got := BucketFor(100, 50, 500); got != Buckets[0] {
		t.Errorf("inverted window gave %d, want the finest bucket", got)
	}
	// A nonsense budget must still produce something usable rather than
	// dividing by zero.
	if got := BucketFor(0, 86400, 0); got <= 0 {
		t.Errorf("zero budget gave %d", got)
	}
	// Ten years at 500 points needs a bucket longer than the table.
	if got := BucketFor(0, 10*365*86400, 500); got <= 0 {
		t.Errorf("very long window gave %d", got)
	}
}

func TestValidate(t *testing.T) {
	ok := domain.MetricQuery{Hostname: "h", Description: "s", From: 1, To: 2}
	if err := Validate(ok); err != nil {
		t.Errorf("a complete query was rejected: %v", err)
	}

	cases := map[string]domain.MetricQuery{
		"no host":         {Description: "s"},
		"no service":      {Hostname: "h"},
		"inverted window": {Hostname: "h", Description: "s", From: 10, To: 5},
	}
	for name, q := range cases {
		t.Run(name, func(t *testing.T) {
			if err := Validate(q); err == nil {
				t.Error("want an error, got nil")
			}
		})
	}
}
