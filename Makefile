.PHONY: all

all: clean
	cd log-producer && zip -r ../dist/log-producer.zip .
	cd e2e && go run .
	# export TESTCONTAINERS_RYUK_DISABLED=true; cd e2e && go run .

stop:
	docker stop $$(docker ps -aq) || true

clean: stop
	cd dist && rm -rf *
	docker rm $$(docker ps -aq) || true
