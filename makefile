test:
	go test -v .\server\internal\service\ .\server\internal\db\

run:
	docker-compose up --build server --attach server --attach client