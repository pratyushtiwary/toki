run-test-pipeline:
	go run . ./examples/configs/pipeline.test.json -p ./examples/configs/project.test.json

unit-test:
	go test -coverpkg=./... -coverprofile=coverage.cov -v ./...

coverage: unit-test
	go tool cover -html=coverage.cov