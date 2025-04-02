PHONY: docker-build logs

docker-build:
	@echo "Starting up the application..."
	@docker-compose -f docker-compose.yml up -d --build

logs:
	@echo "Fetching logs..."
	@docker-compose -f docker-compose.yml logs -f