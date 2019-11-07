all:
	go generate .

docker:
	docker build -t webchat .

publish:
	docker tag webchat sigmonsays/webchat:latest
	docker push sigmonsays/webchat:latest
