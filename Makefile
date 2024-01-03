# Check if the 'plugin' variable is set
validate_plugin:
ifndef plugin
	$(error "The 'plugin' variable is missing. Usage: make build plugin=<plugin_name>")
endif

build: validate_plugin

	# Create a new directory for the build process
	mkdir -p render

	# Copy the entire source tree, excluding .git directory, into the new directory
	rsync -a --exclude='.git' . render/ >/dev/null 2>&1

	# Change to the new directory to perform operations
	cd render && \
	go run generate/generator.go templates . $(plugin) $(plugin_github_url) && \
	go mod tidy && \
	$(MAKE) -f out/Makefile build

	# Clean up the render directory
	rm -rf render

	# Note: The render directory will contain the full code tree with changes, 
	# binaries will be copied to /usr/local/bin, and then render will be deleted
