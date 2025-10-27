.PHONY: build deploy

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go
	zip function.zip bootstrap

deploy: build
	aws lambda update-function-code \
		--function-name hw7-lambda \
		--zip-file fileb://function.zip

clean:
	rm -f bootstrap function.zip