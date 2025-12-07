run-test-pipeline:
	go run . ./examples/configs/pipeline.test.json -p ./examples/configs/project.test.json

test:
	go test -coverpkg=./... -coverprofile=coverage.txt -v ./...

coverage: test
	go tool cover -html=coverage.txt