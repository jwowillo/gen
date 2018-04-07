all: generate gen_server

generate:
	@echo 'generating'
	go generate

gen_server:
	@echo 'making gen_server'
	$(call go,gen_server)
	@echo

define go
	cd cmd/$(1) && go get && go install
endef
