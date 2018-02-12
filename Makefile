SERVICE_NAME   = go-event-sourcing-excercise
IMAGE_NAME     ?= wwgberlin/go-event-sourcing-excercise

up:
	docker-compose up --build -d 

down:
	docker-compose down -v --remove-orphans

logs:
	docker-compose logs -f service

restart: down up
