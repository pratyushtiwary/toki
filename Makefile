run-test-pipeline:
	go run . ./examples/configs/pipeline.test.json -p ./examples/configs/project.test.json

test:
	go test -coverpkg=$$(go list ./... | grep -v /mocks | grep -v /testutils | tr '\n' ',') -coverprofile=coverage.txt ./...

test-verbose:
	go test -coverpkg=$$(go list ./... | grep -v /mocks | grep -v /testutils | tr '\n' ',') -coverprofile=coverage.txt -v ./...

coverage: test
	go tool cover -html=coverage.txt