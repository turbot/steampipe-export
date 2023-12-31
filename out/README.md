# Steampipe Table Dump aws_dump

## Prerequisites
- A build of SQLite that supports extensions (default `brew` install has extensions disabled).
- A build of the aws_dump (default `brew` install aws_dump).

## Usage

### Command Line Interface

The `aws_dump` tool is used from the command line and accepts various options and arguments. Here's an example of how to use it:

```bash
aws_dump [flags] <table_name>
```

**Flags**

* `config`: Specifies the configuration file for the tool. You can provide a file path to load configuration settings.
* `limit`: Sets a limit on the number of rows to retrieve. Useful when you want to restrict the amount of data fetched.
* `select`: Lets you specify the columns you want to display in the output. You can provide a comma-separated list of column names.
* `where`: Allows you to define a WHERE clause to filter the data you want to query. For example, you can filter based on specific conditions.

## Configuration
If you require [configuration](https://hub.steampipe.io/plugins/turbot/aws#configuration) for the extension, you need to set this prior to loading the extension.

## Installation

- Copy the binary `aws_dump` to a directory of choice.
- Run `make plugin_alias="aws" plugin_github_url="github.com/turbot/steampipe-plugin-aws"

## Table Docs

Please refer to the [Table Documentation](https://hub.steampipe.io/plugins/turbot/aws/tables).