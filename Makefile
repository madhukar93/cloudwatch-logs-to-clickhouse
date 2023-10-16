.PHONY: all clean stop log-producer

all: clean log-producer
	cd e2e && LOCALSTACK_API_KEY=$$(bw get password localstack-api-key) go run .
	# export TESTCONTAINERS_RYUK_DISABLED=true; cd e2e && go run .

log-producer:
	cd log-producer && zip -r ../dist/log-producer.zip .

stop:
	docker stop $$(docker ps -aq) || true

clean: stop
	cd dist && rm -rf *
	docker rm $$(docker ps -aq) || true
