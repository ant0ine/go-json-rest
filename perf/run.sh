#!/bin/bash

VERSION=`git log -1 --pretty=format:"%h-%ad" --date=short`

cd ../rest && go test -c && ./rest.test -test.bench="BenchmarkCompression" -test.cpuprofile="cpu.prof"
cd ../rest && go tool pprof --text  rest.test cpu.prof > ../perf/pprof/cpu-$VERSION.txt
cd ../rest && rm -f rest.test cpu.prof

rm -f perf/bench/bench-$VERSION.txt
cd ../rest && go test -bench=. >> ../perf/bench/bench-$VERSION.txt
cd ../rest && go test -bench=. >> ../perf/bench/bench-$VERSION.txt
cd ../rest && go test -bench=. >> ../perf/bench/bench-$VERSION.txt
