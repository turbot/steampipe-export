# Steampipe Exporter

A Steampipe exporter fetches data from cloud services and APIs. Each exporter is a standalone binary that extracts data using a Steampipe plugin. These repository enables you to build an exporter derived from a [Steampipe plugin](https://hub.steampipe.io/plugins).

## Getting Started

If you just want to acquire and run the binary for an exporter, you can download an installer from [Steampipe downloads](https://steampipe.io/downloads). See the [installation docs](https://turbot.com/docs/steampipe_export/install) for details. 

To build an exporter, use the provided `Makefile`. For example, to build the AWS exporter, run the following command to build the tool. It lands in your current directory by default, or elsewhere if you override. 

```bash
make plugin=aws plugin_github_url=github.com/turbot/steampipe-plugin-aws
```

## Prerequisites

- [Golang](https://golang.org/doc/install) Version 1.21 or higher.

## Usage

`steampipe_export_github -h`

```bash
Export data using the github plugin.

Find detailed usage information including table names, column names, and
examples at the Steampipe Hub: https://hub.steampipe.io/plugins/turbot/github

Usage:
  steampipe_export_github TABLE_NAME [flags]

Flags:
      --config string       Config file data
  -h, --help                help for steampipe_export_github
      --limit int           Limit data
      --output string       Output format: csv, json or jsonl (default "csv")
      --select strings      Column data to display
      --where stringArray   where clause data
```

## Examples

### Export EC2 instances using an AWS profile

```bash
./steampipe_export_aws aws_ec2_instance \
  --config='profile="dundermifflin"'
```

### Filter to running instances

```bash
./steampipe_export_aws aws_ec2_instance \
  --config='profile="dundermifflin"' \
  --where="instance_state='running'"
```

### Select a subset of columns

```bash
./steampipe_export_aws aws_ec2_instance \
  --config 'profile="dundermifflin"' \
  --where "instance_state='running'" \
  --select "arn,instance_state"
```

### Limit results

```bash
./steampipe_export_aws aws_ec2_instance \
  --config 'profile="dundermifflin"' \
  --where "instance_state='running'" \
  --select "arn,instance_state" \
  --limit 10
```


## Contributing
If you would like to contribute to this project, please open an issue or create a pull request. We welcome any improvements or bug fixes. Contributions are subject to the [Apache-2.0](https://opensource.org/license/apache-2-0/) license.


