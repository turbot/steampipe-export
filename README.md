# Steampipe Export

A family of export tools, each derived from a [Steampipe plugin](https://hub.steampipe.io/plugins), that fetch data from cloud services and APIs.

## Getting Started

You can use an installer that enables you to choose a plugin and download the export tool for that plugin. See the [installation docs](https://turbot.com/docs/steampipe_export/install) for details. 

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

## Developing

To build an export tool, use the provided `Makefile`. For example, to build the AWS tool, run the following command to build the tool. It lands in `/usr/local/bin` by default, or elsewhere if you override using the `OUTPUT_DIR` environment variable.

```bash
make build plugin=aws
```

## Prerequisites

- [Golang](https://golang.org/doc/install) Version 1.21 or higher.

## Contributing
If you would like to contribute to this project, please open an issue or create a pull request. We welcome any improvements or bug fixes. Contributions are subject to the [Apache-2.0](https://opensource.org/license/apache-2-0/) license.
