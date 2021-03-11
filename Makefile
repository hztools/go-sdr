# {{{ Copyright (c) Paul R. Tagliamonte <paul@k3xec.com>, 2020
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE. }}}

GO := $(shell which go)

SUPPORTED_BUILTAGS = \
	sdr.nosimd \
	sdr.experimental

check: test lint

clean:
	rm -vf benchmark.*

test: $(addprefix test_, $(SUPPORTED_BUILTAGS))
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) test -race ./...

coverage: test
	$(GO) tool cover -html=coverage.out

test_%:
	$(GO) test -tags=$* ./...
	$(GO) test -tags=$* -race ./...

test-simd-helpers:
	@echo "================= SIMD Helper header test suite ================="
	$(GO) test -v -tags=sdr.nosimd,sdr.simdtest \
		-run=.*SimdTest.* \
		./internal/simd

benchmark:
	@echo "================= Benchmark without SIMD ================="
	$(GO) test -count 5 -tags=sdr.nosimd -bench=. ./... | tee benchmark.nosimd
	@echo "================= Benchmark with SIMD ================="
	$(GO) test -count 5                  -bench=. ./... | tee benchmark.default
	@echo "================= Benchstat ================="
	# golang.org/x/perf/cmd/benchstat
	benchstat benchmark.nosimd benchmark.default


lint:
	@echo "================= golint ================="
	golint ./...


.PHONY: all test benchmark lint

# vim: foldmethod=marker
