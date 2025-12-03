build-app-image:
	./build_app.sh $(CMD_NAME) $(IMAGE_NAME)

build-http-server:
	$(MAKE) build-app-image CMD_NAME=http-server IMAGE_NAME=auth_service_http

run-service:
	docker compose up -d --no-deps $(SERVICE_NAME)

run-db:
	$(MAKE) run-service SERVICE_NAME=auth_service_db

run-db-migration:
	$(MAKE) run-service SERVICE_NAME=auth_service_db_migration

run-http-server: build-http-server
	$(MAKE) run-service SERVICE_NAME=auth_service_http

stop-all:
	docker compose down
