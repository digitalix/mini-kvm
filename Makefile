
export IMAGE="mkvm"

build:
	@docker buildx build --output type=local,dest=./out --platform linux/arm64 -t $(IMAGE):$(shell date +'%g%j%H%M%S') -t $(IMAGE):latest -f Dockerfile .

deploy: build
	@scp out/service patryk@192.168.1.89:/home/patryk/

run: deploy
	@ssh patryk@192.168.1.89 "./service"