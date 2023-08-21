IMAGE_NAME = "argocd-bot"
build:
	docker build -t $(IMAGE_NAME) .

run:
	GIT_PRIVATE_KEY=$(cat ~/.ssh/id_rsa) docker-compose run --rm test-runner