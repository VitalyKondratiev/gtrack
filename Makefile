build:
	GOOS=windows GOARCH=amd64 go build . && GOOS=linux GOARCH=amd64 go build . && GOOS=darwin GOARCH=amd64 go build -o gtrack.app
	tar --transform='flags=r;s|abc.xml|abc.xml_v9|' -czf gtrack.tar.gz gtrack.app && rm gtrack.app
clean:
	rm gtrack*