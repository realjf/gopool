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
