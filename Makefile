.PHONY: test-coverage sonar-up sonar-scan sonar-check sonar-down

test-coverage:
	go test -count=1 -covermode=atomic -coverprofile=coverage.out ./...
	@coverage=$$(go tool cover -func=coverage.out | awk '/^total:/ {gsub("%", "", $$3); print $$3}'); \
		awk -v coverage="$$coverage" 'BEGIN { if (coverage < 80) { printf "Coverage %.1f%% is below 80%%\n", coverage; exit 1 } printf "Coverage %.1f%% meets 80%%\n", coverage }'

sonar-up:
	docker compose -f docker-compose.sonar.yml up -d sonarqube

sonar-scan:
	docker compose -f docker-compose.sonar.yml run --rm scanner

sonar-check: test-coverage sonar-scan

sonar-down:
	docker compose -f docker-compose.sonar.yml down
