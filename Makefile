all: build

validate_plugin:
ifndef plugin
	$(error "'plugin' is missing. Usage: make build plugin=<PLUGIN_NAME> plugin_github_url=<github.com/OWNER/steampipe-plugin-<PLUGIN_NAME>")
endif

ifndef plugin_github_url
	$(error "'github_plugin_url' is missing. Usage: make build plugin=<PLUGIN_NAME> plugin_github_url=<github.com/OWNER/steampipe-plugin-<PLUGIN_NAME>")
endif

# Check for output directory
output_dir ?= $(shell read -p "Enter output directory (default /usr/local/bin): " dir; echo $$dir)

build: validate_plugin
	go run generate/generator.go templates . $(plugin) $(plugin_github_url)
	go mod tidy
	make -f out/Makefile build OUTPUT_DIR=$(output_dir)

