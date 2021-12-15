build:
	GOOS=windows GOARCH=amd64 go build . && go build . && GOOS=darwin GOARCH=amd64 go build -o gtrack.app
clean:
	rm gtrack*