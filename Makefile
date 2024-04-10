.PHONY: build
build:
	GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o bootstrap main.go && \
	zip -j manage_subscriptions.zip bootstrap && \
	rm bootstrap

clean:
	rm manage_subscriptions.zip