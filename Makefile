.PHONY: all

all:
	zip -r bin/log-producer.zip ./log-producer
	cd e2e && go run .
	# export TESTCONTAINERS_RYUK_DISABLED=true; cd e2e && go run .
