.PHONY: gen

gen:
	goa gen github.com/flaccid/google-keep-clone/backend/design
	go run ./tools/codegen
	cp gen/http/openapi3.yaml cmd/keep_server/openapi3.yaml
