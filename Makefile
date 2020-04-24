build:
	docker-compose build

run:
	docker-compose up -d

stop:
	docker-compose down -v

test:
	docker-compose run go go test -v ./...
	docker-compose down

