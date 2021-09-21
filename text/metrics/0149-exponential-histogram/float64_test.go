package histogram_test

import (
	"fmt"
	"math/rand"
	"testing"

	histogram "github.com/open-telemetry/oteps/text/metrics/0149"

	"github.com/stretchr/testify/assert"
)

func TestScalb(t *testing.T) {
	assert.Equal(t, 2.0, histogram.Scalb(1, 1))
	assert.Equal(t, 0.5, histogram.Scalb(1, -1))
	assert.Equal(t, -2.0, histogram.Scalb(-1, 1))
	assert.Equal(t, -0.5, histogram.Scalb(-1, -1))
}

func testCompatibility(t *testing.T, histoScale, testScale int) {
	src := rand.New(rand.NewSource(54979))
	t.Run(fmt.Sprintf("compat_%d_%d", histoScale, testScale), func(t *testing.T) {
		const trials = 1e5

		ltm := histogram.NewLookupTableMapping(histoScale)
		lgm := histogram.NewLogarithmMapping(histoScale)

		for i := 0; i < trials; i++ {
			v := histogram.Scalb(1+src.Float64(), testScale)

			lti := ltm.MapToIndex(v)
			lgi := lgm.MapToIndex(v)

			assert.Equal(t, lti, lgi)
		}

		size := int64(1) << histoScale
		additional := int64(testScale) * size

		for i := int64(0); i < size; i++ {
			ltb := ltm.LowerBoundary(i + additional)
			lgb := lgm.LowerBoundary(i + additional)

			assert.InEpsilon(t, ltb, lgb, 0.000001, "hs %v ts %v sz %v add %v index %v ltb %v lgb %v", histoScale, testScale, size, additional, i+additional, ltb, lgb)
		}
	})
}

func TestCompat0(t *testing.T) {
	for scale := 3; scale <= 10; scale++ {
		for tscale := -1; tscale <= 1; tscale++ {
			testCompatibility(t, scale, tscale)
		}
	}
}
