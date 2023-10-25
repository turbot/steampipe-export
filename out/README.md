# AWS Table Dump

This tool dumps data from Steampipe plugins. With aws Table Dump at your disposal, you can effortlessly retrieve data from your cloud APIs with exceptional ease and efficiency.

## Prerequisites
- A build of the aws_dump (default `brew` install aws_dump).

## Usage

### Command Line Interface

The `aws_dump` tool is used from the command line and accepts various options and arguments. Here's an example of how to use it:

```bash
aws_dump [flags] <table_name>
```

**Flags**

* `config`: lets you set the configuration options that are supported by the underlying [Steampipe plugin](https://hub.steampipe.io/plugins/turbot/aws/configuration).
* `limit`: Sets a limit on the number of rows to retrieve. Useful when you want to restrict the amount of data fetched.
* `select`: Lets you specify the columns you want to display in the output. You can provide a comma-separated list of column names.
* `where`: Allows you to define a WHERE clause to filter the data you want to query. For example, you can filter based on specific conditions.

## Installation

- Copy the binary `aws_dump` to a directory of choice.
- Run `make plugin_alias="aws" plugin_github_url="github.com/turbot/steampipe-plugin-aws"

## Table Docs

Please refer to the [Table Documentation](https://hub.steampipe.io/plugins/turbot/aws/tables).

## Contributing

If you would like to contribute to this project, please open an issue or create a pull request. We welcome any improvements or bug fixes.

## License

This project is licensed under the [Apache 2.0 open source license](https://github.com/turbot/steampipe-table-dump/blob/main/LICENSE) - see the LICENSE file for details.