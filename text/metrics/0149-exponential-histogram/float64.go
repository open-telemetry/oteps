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

package histogram

import "math"

const (
	// MantissaWidth is the size of an IEEE 754 double-precision
	// floating-point Mantissa.
	MantissaWidth = 52
	// ExponentWidth is the size of an IEEE 754 double-precision
	// floating-point exponent.
	ExponentWidth = 11

	// MantissaOnes is MantissaWidth 1 bits
	MantissaOnes = 1<<MantissaWidth - 1

	// ExponentBias is the exponent bias specified for encoding
	// the IEEE 754 double-precision floating point exponent.
	ExponentBias = 1<<(ExponentWidth-1) - 1

	// ExponentMask are set to 1 for the bits of an IEEE 754
	// floating point exponent (as distinct from the Mantissa and
	// sign.
	ExponentMask = ((1 << ExponentWidth) - 1) << MantissaWidth
)

// java.lang.Math.scalb(float f, int scaleFactor) returns f x
// 2**scaleFactor, rounded as if performed by a single correctly
// rounded floating-point multiply to a member of the double value set.
func scalb(f float64, sf int) float64 {
	valueBits := math.Float64bits(f)

	mantissa := MantissaOnes & valueBits

	exponent := int64((ExponentMask & valueBits) >> MantissaWidth)
	exponent += int64(sf)

	return math.Float64frombits(uint64(exponent<<MantissaWidth) | mantissa)
}
