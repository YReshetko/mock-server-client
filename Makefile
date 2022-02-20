MOCKSERVER_CONTAINER_ID="$(shell docker ps -a --filter "name=mockserver" --format "{{.ID}}")"

.PHONY: docker-up
docker-up:
	docker-compose up -d mockserver

.PHONY: docker-down
docker-down:
	docker-compose stop && docker-compose down #&& docker rmi -f mockserver/mockserver:latest

.PHONY: docker-logs
docker-logs:
	docker logs $(MOCKSERVER_CONTAINER_ID)

.PHONY: docker-rerun
docker-rerun: docker-down docker-up
