.PHONY: test
test:
# 运行所有测试
# go test
# 运行所有压力测试
# go test -test.bench=".*"
# 运行单个文件
# go test pool_test.go pool.go task.go worker.go -test.bench=".*"
# 只运行单个方法
# go test -v -test.run BenchmarkPoolRun -test.bench=".*"
	@echo 'run test...'
	@go test -v ./...


push:
	@git add -A && git commit -m "update" && git push origin master


# 运行基准测试
.PHONY: bench
bench:
	@go test -v -bench=. -run=none

# 运行数据竞争检测
.PHONY: race
race:
	@go test -race -v ./... -run=.*


.PHONY: bench race
all:
	@go test -race -run ^TestPoolRun$ github.com/realjf/gopool_test -count=1 -v

# lint
.PHONY: lint
lint:
	@golangci-lint run -v ./...
