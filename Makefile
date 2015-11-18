.PHONY: deps
SHELL=/bin/bash
MAKEFILE_DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
CURRENT_DIR := $(notdir $(patsubst %/,%,$(dir $(MAKEFILE_DIR))))
DIR=$(MAKEFILE_DIR)
TESTDIR=$(DIR)/test
package_dir=dispatch
local_depends:=$(wildcard $(package_dir)/*.go)
libargs=utilities.go common.go
targets:=$(patsubst %.go,bin/%,$(filter-out $(libargs),$(wildcard *.go)))
yaml2json=$(GOPATH)/bin/yaml2json
json2yaml=$(GOPATH)/bin/json2yaml

libdep=github.com/go-logfmt/logfmt\
 github.com/go-kit/kit/{endpoint,log,transport/http} \
 github.com/davidwalter0/logger \
 github.com/jehiah/go-strftime \
 github.com/hashicorp/consul/api \
 github.com/mattn/go-sqlite3 \
 github.com/jinzhu/gorm

all:
	@printf "\n--------------------------------\n"
	@printf "Running in abs directory\n    $(MAKEFILE_DIR)\n"
	@printf "The subdirectory is $(notdir $(patsubst %/,%,$(dir $(MAKEFILE_DIR))))\n"
	@printf "\n--------------------------------\n"
	@printf "make targets init to initialize godeps, get, save, test and build\n"

build: $(targets) 

deps: $(package_dir)/.dep .dep 

.dep: $(targets)
	touch .dep

$(package_dir)/.dep: $(local_depends)
	(								\
		cd $(package_dir);					\
		GO15VENDOREXPERIMENT=1 GOPATH=${GOPATH}			\
		   $(GOPATH)/bin/godep go build -a -ldflags '-s';	\
		GO15VENDOREXPERIMENT=1 GOPATH=${GOPATH}			\
		   $(GOPATH)/bin/godep go install -ldflags "-s"		\
	)
	touch $(package_dir)/.dep

%: bin/% $(libargs)

BUILDARGS:=$(shell rm bin/api-service; echo \-X main.Build="$$(date -u +%Y.%m.%d.%H.%M.%S.%:::z)" \-X main.Commit="\$$(git log --format=%hash-%aI -n1)")

ARGS:=-tags netgo -ldflags "-s -w $(BUILDARGS)"
bin/%: %.go $(libargs) $(package_dir)/.dep $(local_depends)
	@echo "Building via % rule for $@ from $<"
	@mkdir -p bin;
	echo build=BUILDARGS  build=$(BUILDARGS) args=$(ARGS)
	GO15VENDOREXPERIMENT=1 CGO_ENABLED=0 GOPATH=${GOPATH} \
		$(GOPATH)/bin/godep go build $(ARGS) -o $@ $< $(libargs)
#		$(GOPATH)/bin/godep go build -a -ldflags '-s' -o $@ $< $(libargs)
version: bin/api-service
	bin/api-service --version
#	bin/api-service --verbose --debug

init: get save

get: 
	GO15VENDOREXPERIMENT=1 GOPATH=${GOPATH} $(GOPATH)/bin/godep get $(libdep)
save:
	GO15VENDOREXPERIMENT=1 GOPATH=${GOPATH} $(GOPATH)/bin/godep save
test: 
	echo No unit tests written, see transform package

test-driver: .dep
	-bin/driver --file           $(TESTDIR)/invalid.yaml
	-bin/driver --verbose --file $(TESTDIR)/unit2.json
	-bin/driver --verbose --file $(TESTDIR)/unit2.yaml
	-bin/driver --file 	     $(TESTDIR)/test-unit-example.json
	-bin/driver --file 	     $(TESTDIR)/unit2.json
	-bin/driver --file 	     $(TESTDIR)/unit2-ok.json
	-bin/driver --file 	     $(TESTDIR)/check-echo.yaml
	-bin/driver --file 	     $(TESTDIR)/check-tcp.yaml
	-bin/driver --file 	     $(TESTDIR)/k8s-health.yaml
	-bin/driver --file 	     $(TESTDIR)/check-echo-fail.yaml
	-bin/driver --file 	     $(TESTDIR)/consul-cmd.yaml

test-file-list=$(shell echo check-echo.yaml check-tcp.yaml k8s-health.yaml	\
	unit2.yaml unit.yaml k8s-health.json					\
	unit2.json unit2-ok.json unit.json					\
	check-echo-fail.yaml cmd-{echo,docker}.{json,yaml}			\
	consul-cmd.yaml consul-cmd.json)

test-files=$(patsubst %,test/%,$(test-file-list))

apiVersion=v1

test-echo: .dep
	port=1234; $(DIR)/bin/api-service --port=$${port} --debug --verbose > $(DIR)/log.tmp 2>&1 & pid=$$!; \
	until curl localhost:$${port} > /dev/null 2>&1 ; do sleep 1; done;\
	for file in test/cmd-echo.yaml; do\
	  if [[ $${file##*.} == "yaml" ]]; then	\
	      curl --silent -XPOST -d"$$($(DIR)/bin/yaml2json --compress < $${file})" localhost:$${port}/api/$(apiVersion)/exec;	\
	  fi;	\
	done
#; kill $${pid}

test-echo-jq: .dep
	port=1234; $(DIR)/bin/api-service --port=$${port} --debug --verbose > $(DIR)/log.tmp 2>&1 & pid=$$!; \
	until curl localhost:$${port} > /dev/null 2>&1 ; do sleep 1; done;\
	for file in test/cmd-echo.yaml; do\
	  if [[ $${file##*.} == "yaml" ]]; then	\
	      curl --silent -XPOST -d"$$($(DIR)/bin/yaml2json --compress < $${file})" localhost:$${port}/api/$(apiVersion)/exec |\
		 $(DIR)/bin/jq '.spec.status[]|{exit:.exit, when: .timestamp, name:.name}'|cat;  \
	  fi;	\
	done
# ; kill $${pid} make wil kill the children processes. . . ?
#	done| ${DIR}/bin/jq --raw-output --compact-output '.spec.status[]|{exit:.exit, when: .timestamp, name:.name}'|cat

test-api-service: .dep test-api-service-minimal
	@echo; if((0)); then														\
	for file in $(test-files) ; do													\
	  if [[ $${file##*.} == "yaml" ]]; then												\
	      echo "curl --silent -XPOST -d'$$($(DIR)/bin/yaml2json --compress < $${file})' localhost:9990/api/${apiVersion}/exec";	\
	    else															\
	      echo "curl --silent -XPOST -d@$${file} localhost:9990/api/${apiVersion}/exec";					\
	  fi;																\
	done; 		fi;

test-api-service-jq:
	echo;														\
	echo $(test-files);											\
	for file in $(test-files) ; do											\
	  if [[ $${file##*.} == "yaml" ]]; then										\
	      curl --silent -XPOST -d"$$($(DIR)/bin/yaml2json < $${file})" localhost:9990/api/${apiVersion}/exec;	\
	    else													\
	      curl --silent -XPOST -d@$${file} localhost:9990/api/$${apiVersion}/exec;				\
	  fi;														\
	done| ${DIR}/bin/jq --raw-output --compact-output '.spec.status[]|{exit:.exit, when: .timestamp, name:.name}'|cat

test-api-service-minimal:
	for file in $(test-files) ; do											\
	  if [[ $${file##*.} == "yaml" ]]; then										\
	      curl --silent -XPOST -d"$$($(DIR)/bin/yaml2json < $${file})" localhost:9990/api/${apiVersion}/exec;	\
	    else													\
	      curl --silent -XPOST -d@$${file} localhost:9990/api/${apiVersion}/exec;				\
	  fi;														\
	done

clean:
	@echo cleaning up temporary files
	@echo rm -f $(targets)
	@rm -f $(targets) .dep $(package_dir)/.dep

