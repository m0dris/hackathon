up:
	@docker compose up -d --build

down:
	@docker compose down

restart:
	@docker compose restart $(svc)

rebuild:
	@if [ -z "$(svc)" ]; then \
		echo "Отсутствует необходимый параметр svc. Пример: make rebuild svc=golang-service"; \
		exit 1; \
	fi; \
	docker compose up -d --build --no-deps $(svc)

deps-up:
	@docker compose up -d --build postgres api

go-run:
	@docker compose stop golang-service
	@$(MAKE) -C golang-service hackathon-run

proto-gen:
	@$(MAKE) -C golang-service proto-gen

clean:
	@read -p "Удалить контейнеры и volume с данными БД? Данные будут потеряны. [y/N]: " ans; \
	if [ "$$ans" = "y" ]; then \
		docker compose down -v --remove-orphans && \
		echo "Стек и данные БД удалены"; \
	else \
		echo "Очистка отменена"; \
	fi
