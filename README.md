# AWS Dump 

The AWS Dump can effortlessly retrieve data from your Steampipe plugins with exceptional ease and efficiency.

Utilize the AWS Dump to effortlessly retrieve AWS data. This extension uses the [Steampipe Plugin](https://github.com/turbot/steampipe-plugin-aws) for its operation, offering a seamless connection between Steampipe and AWS.

## Quick start

Copy the binary `aws_dump` to a directory of choice

Build, which automatically installs the new version to your directory of choice:

```
make aws_dump
```

Run a query:

```
aws_dump aws_account
```



## Getting Started

To get started with the Steampipe Data Dump tool, you'll need to build and install it. You can do so using the provided `Makefile`. Run the following command to build the tool and install it in the specified directory (default is `/usr/local/bin`):

```bash
make
```

## Prerequisites

- [Golang](https://golang.org/doc/install) Version 1.21 or higher.

## Usage

### Command Line Interface

The `spdump` tool is used from the command line and accepts various options and arguments. Here's an example of how to use it:

```bash
spdump [flags] <table_name>
```

**Flags**

* `config`: Specifies the configuration file for the tool. You can provide a file path to load configuration settings.
* `limit`: Sets a limit on the number of rows to retrieve. Useful when you want to restrict the amount of data fetched.
* `select`: Lets you specify the columns you want to display in the output. You can provide a comma-separated list of column names.
* `where`: Allows you to define a WHERE clause to filter the data you want to query. For example, you can filter based on specific conditions.

Example

```bash
spdump aws_s3_bucket --select name,arn,region,account_id --limit 100
```

## Quick start

Copy the binary `aws_dump` to a directory of choice

Build, which automatically installs the new version to your directory of choice:

```
make aws_dump
```

Run a query:

```
aws_dump aws_account
```

## Contributing
If you would like to contribute to this project, please open an issue or create a pull request. We welcome any improvements or bug fixes.

## License
This project is licensed under the [Apache 2.0 open source license](https://github.com/turbot/steampipe-table-dump/blob/main/LICENSE) - see the LICENSE file for details.

