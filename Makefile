TAG := 1.0

all:
	go generate .

docker:
	docker build -t webchat .

publish:
	docker tag webchat sigmonsays/webchat:$(TAG)
	docker push sigmonsays/webchat:$(TAG)
