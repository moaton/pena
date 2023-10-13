test:
	go test -v .\server\internal\service\

run:
	docker-compose up --build server --attach server --attach client