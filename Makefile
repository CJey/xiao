APPNAME := xiao

help:
	@echo "\033[m┌─────────────────────────────────────────────────"
	@echo "\033[m│ * AppName:  \033[34m$(APPNAME)"
	@echo "\033[m│ * MakeItem: \033[34m\c"
	@make -qp | sed -nr 's/^([a-zA-Z][a-zA-Z0-9_-]*):([^=]|$$)/\1/p' | grep -v '^Makefile$$' | grep -v '^help$$' | sort -u | tr '\n' ' ' && echo
	@echo "\033[m└─────────────────────────────────────────────────"

# 只留module行重新tidy，全部依赖升到最新；失败时还原go.mod
tidy:
	cp go.mod go.mod.bak && grep -m1 '^module ' go.mod.bak > go.mod && \
		if go mod tidy; then rm -f go.mod.bak; else mv go.mod.bak go.mod; exit 1; fi

vet:
	CGO_ENABLED=0 go fix ./...
	go fmt ./...
	CGO_ENABLED=0 go vet ./...
