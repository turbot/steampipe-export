# Steampipe Export

A family of export tools, each derived from a [Steampipe plugin](https://hub.steampipe.io/plugins), that fetch data from cloud services and APIs.

## Getting Started

You can use an installer that enables you to choose a plugin and download the export tool for that plugin.

[Installation guide →](https://steampipe.io/docs/steampipe_export/install)

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

## Open Source & Contributing

This repository is published under the [Apache 2.0](https://www.apache.org/licenses/LICENSE-2.0). 

Please see our [code of conduct](https://github.com/turbot/.github/blob/main/CODE_OF_CONDUCT.md). We look forward to collaborating with you!

[Steampipe](https://steampipe.io) is a product produced exclusively by [Turbot HQ, Inc](https://turbot.com). It is distributed under our commercial terms. Others are allowed to make their own distribution of the software, but cannot use any of the Turbot trademarks, cloud services, etc. You can learn more in our [Open Source FAQ](https://turbot.com/open-source).
