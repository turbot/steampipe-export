# AWS Table Dump

Utilize the AWS Dump to effortlessly retrieve AWS data. This extension uses the [Steampipe Plugin](https://github.com/turbot/steampipe-plugin-aws) for its operation, offering a seamless connection between Steampipe and AWS.

## Installation

- Copy a build of the aws_dump (default `brew` install aws_dump).
- Run `make plugin_alias="aws" plugin_github_url="github.com/turbot/steampipe-plugin-aws`

## Quick start

Copy the binary `aws_dump` to a directory of choice

Build, which automatically installs the new version to your directory of choice:

```
make make plugin_alias="aws" plugin_github_url="github.com/turbot/steampipe-plugin-aws
```

Run a query:

```
aws_dump aws_account
```

Run a query with different flags:

```shell
spdump aws_s3_bucket --select name,arn,region,account_id --limit 100 
account_id,arn,name,region
632902152528,arn:aws:s3:::aws-glue-assets-632902152528-us-east-1,aws-glue-assets-632902152528-us-east-1,us-east-1
632902152528,arn:aws:s3:::sp-flow-s3bucket-logsink,sp-flow-s3bucket-logsink,us-west-2
632902152528,arn:aws:s3:::aws-security-data-lake-us-east-2-7pthxugfyv6u5uzyd6f3qt0tgd1mlu,aws-security-data-lake-us-east-2-7pthxugfyv6u5uzyd6f3qt0tgd1mlu,us-east-2
632902152528,arn:aws:s3:::cf-templates-yyk2l1c0d6k5-us-east-1,cf-templates-yyk2l1c0d6k5-us-east-1,us-east-1
632902152528,arn:aws:s3:::aws-logs-632902152528-us-east-1,aws-logs-632902152528-us-east-1,us-east-1
```

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

## Table Docs

Please refer to the [Table Documentation](https://hub.steampipe.io/plugins/turbot/aws/tables).

## Contributing

If you would like to contribute to this project, please open an issue or create a pull request. We welcome any improvements or bug fixes.

## License

This project is licensed under the [Apache 2.0 open source license](https://github.com/turbot/steampipe-table-dump/blob/main/LICENSE) - see the LICENSE file for details.