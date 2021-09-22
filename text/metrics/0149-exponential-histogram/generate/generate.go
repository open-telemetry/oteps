// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"math"
	"math/big"
	"math/bits"
	"os"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	histogram "github.com/open-telemetry/oteps/text/metrics/0149"
)

// main prints a table of constants for use in a lookup-table
// implementation of the base2 exponential histogram of OTEP 149.
//
// Note: this has exponential time complexity because the number of
// math/big operations per entry in the table is O(2**scale).
//
// "generate 12" takes around 5 minutes of CPU time.
func main() {
	scale, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Printf("usage: %s scale (an integer)\n", os.Args[1])
		os.Exit(1)
	}

	var (
		// size is 2^scale
		size       = int64(1) << scale
		thresholds = make([]uint64, size)

		// constants
		onef = big.NewFloat(1)
		onei = big.NewInt(1)
	)

	newf := func() *big.Float { return &big.Float{} }
	newi := func() *big.Int { return &big.Int{} }
	pow2 := func(x int) *big.Float {
		return newf().SetMantExp(onef, x)
	}
	toInt64 := func(x *big.Float) *big.Int {
		i, _ := x.SetMode(big.ToZero).Int64()
		return big.NewInt(i)
	}
	ipow := func(b *big.Int, p int64) *big.Int {
		r := onei
		for i := int64(0); i < p; i++ {
			r = newi().Mul(r, b)
		}
		return r
	}
	start := time.Now()

	var finished int64
	var wg sync.WaitGroup

	// Round to a power of two larger than NumCPU
	ncpu := 1 << (64 - bits.LeadingZeros64(uint64(runtime.NumCPU())))
	percpu := len(thresholds) / ncpu
	wg.Add(ncpu)

	go func() {
		// Since this can take a long time to run for large
		// scales, print a progress report.  This assumes the
		// job is not suspended...
		t := time.NewTicker(time.Minute)
		defer t.Stop()

		for {
			select {
			case <-t.C:
				if finished == 0 {
					continue
				}
				elapsed := time.Since(start)
				count := atomic.LoadInt64(&finished)
				os.Stderr.WriteString(fmt.Sprintf("%d @ %s: %.4f%% complete %s remaining...\n",
					count,
					elapsed.Round(time.Minute),
					100*float64(count)/float64(size),
					time.Duration(
						float64(size)*float64(elapsed)/float64(count)-
							float64(elapsed),
					).Round(time.Minute),
				))
			}
		}
	}()

	for cpu := 0; cpu < ncpu; cpu++ {
		go func(cpu int) {
			defer wg.Done()
			for j := 0; j < percpu; j++ {
				position := cpu*percpu + j

				// whereas (position/size) in the range [0, 1),
				//   x = 2^(position/size)
				// falls in the range [1, 2).  Equivalently,
				// calculate 2^position, then square-root scale times.
				x := pow2(position)
				for i := 0; i < scale; i++ {
					x = newf().Sqrt(x)
				}

				// Compute the integer value in the range [2^52, 2^53)
				// which is the 52-bit significand of the IEEE float64
				// as an uint64 value plus 2^52.
				scaled := newf().Mul(x, pow2(52))
				ieeeNormalized := toInt64(scaled) // in the range [2^52, 2^53)

				compareTo, _ := pow2(52*int(size) + position).Int(nil)

				large := ipow(ieeeNormalized, size)

				if large.Cmp(compareTo) < 0 {
					ieeeNormalized = newi().Add(ieeeNormalized, onei)
				}

				thresholds[position] = ieeeNormalized.Uint64() & ((uint64(1) << 52) - 1)

				// Validate that this is the correct result by
				// subtracting one, ensure that the value is less than
				// compareTo.
				sigLessOne := newi().Sub(ieeeNormalized, onei)

				// If (ieeeNormalized-1)^size is greater than or equal to the
				// inclusive lower bound
				if ipow(sigLessOne, size).Cmp(compareTo) >= 0 {
					panic("incorrect result")
				}
				atomic.AddInt64(&finished, 1)
			}
		}(cpu)
	}
	wg.Wait()

	fmt.Printf(`package histogram

var exponentialConstants = [%d]uint64{
`, size)

	for pos, value := range thresholds {
		fmt.Printf("\t0x%012x, // significand(2^(%d/%d) == %.016g)\n",
			value,
			pos,
			size,
			math.Float64frombits((uint64(histogram.ExponentBias)<<histogram.MantissaWidth)+value),
		)
	}
	fmt.Printf(`}
`)
}
