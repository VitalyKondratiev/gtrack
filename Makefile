build:
	GOOS=windows GOARCH=amd64 go build . && go build . && GOOS=darwin GOARCH=386 go build -o gtrack.app
clean:
	rm gtrack*