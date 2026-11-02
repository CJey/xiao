APPNAME := xiao

# 各目标的输出：各步的说明与上一步之间空一行，命令行所给的目标结束时打出总耗时，计自make启动。
# 配方临执行才按先后展开，eval记下已有过一步，本次make的首步之前便不空行
BEGIN := $(shell date +%s%3N)
step = printf '$(if $(STEPPED),\n)\033[1;36m▶ %s\033[0m\n' "$(1)"$(eval STEPPED := 1)
finish = { [ "$@" != "$(lastword $(MAKECMDGOALS))" ] || date +%s%3N | awk -v b=$(BEGIN) -v t="make $@" \
	'{ s = int(($$1 - b) / 100 + 0.5) / 10; printf "\n\033[1;32m✔ %s done in ", t; if (s >= 60) printf "%dm%04.1fs", s / 60, s % 60; else printf "%.1fs", s; print "\033[0m" }'; }

# 各目标的说明取自其上一行的注释，到第一个分号为止
help:
	@printf '\033[1m$(APPNAME)\033[0m: make <target>\n\n'
	@awk '/^# / { c = substr($$0, 3); next } /^[a-zA-Z][a-zA-Z0-9_-]*:([^=]|$$)/ && !/^help:/ { n = $$1; sub(/:.*/, "", n); \
		sub(/；.*/, "", c); printf "  \033[1;36m%-8s\033[0m %s\n", n, c } { c = "" }' Makefile

# 清掉测试与性能分析的产物
clean:
	@$(call step,Clean the products of testing and profiling)
	@rm -f *.log *.test *.out *.prof

# 只留module行重新tidy，全部依赖升到最新；失败时还原go.mod
tidy:
	@$(call step,Tidy go.mod from scratch with every dependency at its latest)
	@cp go.mod go.mod.bak && grep -m1 '^module ' go.mod.bak > go.mod && \
		if go mod tidy; then rm -f go.mod.bak; else mv go.mod.bak go.mod; exit 1; fi
	@$(finish)

# 依次go fix、go fmt与go vet
vet:
	@$(call step,Run go fix)
	@CGO_ENABLED=0 go fix ./...
	@$(call step,Run go fmt)
	@go fmt ./...
	@$(call step,Run go vet)
	@CGO_ENABLED=0 go vet ./...
	@$(finish)
