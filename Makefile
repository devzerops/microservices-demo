.PHONY: help build-experimental start-experimental stop-experimental logs-experimental clean-experimental test-experimental

# Colors for output
BLUE := \033[0;34m
GREEN := \033[0;32m
RED := \033[0;31m
NC := \033[0m # No Color

help: ## Show this help message
	@echo "$(BLUE)Experimental Services - Make Commands$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ""

build-experimental: ## Build all experimental services
	@echo "$(BLUE)Building experimental services...$(NC)"
	docker-compose -f docker-compose-experimental.yml build
	@echo "$(GREEN)Build complete!$(NC)"

start-experimental: ## Start all experimental services
	@echo "$(BLUE)Starting experimental services...$(NC)"
	docker-compose -f docker-compose-experimental.yml up -d
	@echo "$(GREEN)Services started!$(NC)"
	@echo ""
	@echo "$(BLUE)Service URLs:$(NC)"
	@echo "  🌐 API Gateway:     http://localhost:8080"
	@echo "  🎨 Demo Dashboard:  http://localhost:3000"
	@echo "  📷 Visual Search:   http://localhost:8093"
	@echo "  🎮 Gamification:    http://localhost:8094"
	@echo "  📦 Inventory:       http://localhost:8092"
	@echo "  📱 PWA:             http://localhost:8095"
	@echo "  🔍 Search:          http://localhost:8097"
	@echo "  📊 Analytics:       http://localhost:8099"
	@echo ""
	@echo "$(GREEN)🚀 Open http://localhost:3000 for interactive demo!$(NC)"
	@echo "Run '$(GREEN)make logs-experimental$(NC)' to view logs"

stop-experimental: ## Stop all experimental services
	@echo "$(BLUE)Stopping experimental services...$(NC)"
	docker-compose -f docker-compose-experimental.yml down
	@echo "$(GREEN)Services stopped!$(NC)"

restart-experimental: ## Restart all experimental services
	@echo "$(BLUE)Restarting experimental services...$(NC)"
	docker-compose -f docker-compose-experimental.yml restart
	@echo "$(GREEN)Services restarted!$(NC)"

logs-experimental: ## View logs from all experimental services
	docker-compose -f docker-compose-experimental.yml logs -f

logs-visual: ## View Visual Search service logs
	docker-compose -f docker-compose-experimental.yml logs -f visualsearch

logs-gamification: ## View Gamification service logs
	docker-compose -f docker-compose-experimental.yml logs -f gamification

logs-inventory: ## View Inventory service logs
	docker-compose -f docker-compose-experimental.yml logs -f inventory

logs-pwa: ## View PWA service logs
	docker-compose -f docker-compose-experimental.yml logs -f pwa

logs-search: ## View Search service logs
	docker-compose -f docker-compose-experimental.yml logs -f search

logs-analytics: ## View Analytics service logs
	docker-compose -f docker-compose-experimental.yml logs -f analytics

logs-gateway: ## View API Gateway logs
	docker-compose -f docker-compose-experimental.yml logs -f apigateway

logs-demo: ## View Demo Dashboard logs
	docker-compose -f docker-compose-experimental.yml logs -f demo

ps-experimental: ## Show status of all experimental services
	@docker-compose -f docker-compose-experimental.yml ps

health-check: ## Check health of all experimental services
	@echo "$(BLUE)Checking service health...$(NC)"
	@echo ""
	@echo -n "API Gateway:      "; curl -s http://localhost:8080/health > /dev/null 2>&1 && echo "$(GREEN)✓ Healthy$(NC)" || echo "$(RED)✗ Unhealthy$(NC)"
	@echo -n "Demo Dashboard:   "; curl -s http://localhost:3000/ > /dev/null 2>&1 && echo "$(GREEN)✓ Healthy$(NC)" || echo "$(RED)✗ Unhealthy$(NC)"
	@echo -n "Visual Search:    "; curl -s http://localhost:8093/health > /dev/null 2>&1 && echo "$(GREEN)✓ Healthy$(NC)" || echo "$(RED)✗ Unhealthy$(NC)"
	@echo -n "Gamification:     "; curl -s http://localhost:8094/health > /dev/null 2>&1 && echo "$(GREEN)✓ Healthy$(NC)" || echo "$(RED)✗ Unhealthy$(NC)"
	@echo -n "Inventory:        "; curl -s http://localhost:8092/health > /dev/null 2>&1 && echo "$(GREEN)✓ Healthy$(NC)" || echo "$(RED)✗ Unhealthy$(NC)"
	@echo -n "PWA:              "; curl -s http://localhost:8095/health > /dev/null 2>&1 && echo "$(GREEN)✓ Healthy$(NC)" || echo "$(RED)✗ Unhealthy$(NC)"
	@echo -n "Search:           "; curl -s http://localhost:8097/health > /dev/null 2>&1 && echo "$(GREEN)✓ Healthy$(NC)" || echo "$(RED)✗ Unhealthy$(NC)"
	@echo -n "Analytics:        "; curl -s http://localhost:8099/health > /dev/null 2>&1 && echo "$(GREEN)✓ Healthy$(NC)" || echo "$(RED)✗ Unhealthy$(NC)"

clean-experimental: ## Stop and remove all experimental services and volumes
	@echo "$(RED)Cleaning up experimental services...$(NC)"
	docker-compose -f docker-compose-experimental.yml down -v
	@echo "$(GREEN)Cleanup complete!$(NC)"

test-experimental: ## Run integration tests for experimental services
	@echo "$(BLUE)Testing experimental services...$(NC)"
	@echo ""
	@echo "$(BLUE)Testing Visual Search...$(NC)"
	@curl -X GET http://localhost:8093/health || echo "$(RED)Visual Search not responding$(NC)"
	@echo ""
	@echo "$(BLUE)Testing Gamification...$(NC)"
	@curl -X GET http://localhost:8094/health || echo "$(RED)Gamification not responding$(NC)"
	@echo ""
	@echo "$(BLUE)Testing Inventory...$(NC)"
	@curl -X GET http://localhost:8092/health || echo "$(RED)Inventory not responding$(NC)"
	@echo ""
	@echo "$(BLUE)Testing PWA...$(NC)"
	@curl -X GET http://localhost:8095/health || echo "$(RED)PWA not responding$(NC)"
	@echo ""
	@echo "$(GREEN)All tests complete!$(NC)"

rebuild-experimental: ## Rebuild and restart all experimental services
	@echo "$(BLUE)Rebuilding experimental services...$(NC)"
	docker-compose -f docker-compose-experimental.yml down
	docker-compose -f docker-compose-experimental.yml build --no-cache
	docker-compose -f docker-compose-experimental.yml up -d
	@echo "$(GREEN)Rebuild complete!$(NC)"

# Development targets
dev-visual: ## Run Visual Search in development mode
	@echo "$(BLUE)Starting Visual Search in dev mode...$(NC)"
	cd src/visualsearchservice && uvicorn app.main:app --reload --host 0.0.0.0 --port 8093

dev-gamification: ## Run Gamification in development mode
	@echo "$(BLUE)Starting Gamification in dev mode...$(NC)"
	cd src/gamificationservice && go run *.go

dev-inventory: ## Run Inventory in development mode
	@echo "$(BLUE)Starting Inventory in dev mode...$(NC)"
	cd src/inventoryservice && go run *.go

dev-pwa: ## Run PWA in development mode
	@echo "$(BLUE)Starting PWA in dev mode...$(NC)"
	cd src/pwa-service && npm install && npm run dev

dev-search: ## Run Search in development mode
	@echo "$(BLUE)Starting Search in dev mode...$(NC)"
	cd src/searchservice && go run *.go

dev-analytics: ## Run Analytics in development mode
	@echo "$(BLUE)Starting Analytics in dev mode...$(NC)"
	cd src/analyticsservice && go run *.go

dev-gateway: ## Run API Gateway in development mode
	@echo "$(BLUE)Starting API Gateway in dev mode...$(NC)"
	cd src/apigateway && go run *.go

dev-demo: ## Run Demo Dashboard in development mode
	@echo "$(BLUE)Starting Demo Dashboard in dev mode...$(NC)"
	cd src/demo-dashboard && npm install && npm start

# Demo data
demo-data: ## Load demo data into all services
	@echo "$(BLUE)Loading demo data...$(NC)"
	@# Index demo products in Visual Search
	@curl -X POST http://localhost:8093/index \
		-H "Content-Type: application/json" \
		-d '{"products":[{"product_id":"OLJCESPC7Z","name":"Sunglasses","price":19.99,"image_url":"http://example.com/sunglasses.jpg"},{"product_id":"66VCHSJNUP","name":"Tank Top","price":18.99,"image_url":"http://example.com/tank.jpg"}]}' \
		|| echo "$(RED)Failed to index products$(NC)"
	@# Award demo points
	@curl -X POST http://localhost:8094/users/demo-user/points \
		-H "Content-Type: application/json" \
		-d '{"points":100,"action":"signup","reason":"Welcome bonus"}' \
		|| echo "$(RED)Failed to award points$(NC)"
	@echo "$(GREEN)Demo data loaded!$(NC)"

# Documentation
docs: ## Open integration documentation
	@if command -v xdg-open > /dev/null; then \
		xdg-open EXPERIMENTAL_SERVICES_INTEGRATION.md; \
	elif command -v open > /dev/null; then \
		open EXPERIMENTAL_SERVICES_INTEGRATION.md; \
	else \
		echo "Please open EXPERIMENTAL_SERVICES_INTEGRATION.md manually"; \
	fi

open-demo: ## Open demo dashboard in browser
	@echo "$(BLUE)Opening demo dashboard...$(NC)"
	@if command -v xdg-open > /dev/null; then \
		xdg-open http://localhost:3000; \
	elif command -v open > /dev/null; then \
		open http://localhost:3000; \
	else \
		echo "Please open http://localhost:3000 manually"; \
	fi

# Default target
.DEFAULT_GOAL := help
