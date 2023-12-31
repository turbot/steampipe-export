#!/bin/sh

set -e

main() {
  # ANSI escape code variables
  BOLD=$(tput bold)
  NORMAL=$(tput sgr0)

  if ! command -v tar >/dev/null 2>&1; then
    echo "Error: 'tar' is required." 1>&2
    exit 1
  fi

  OS=$(uname -s)
  if [ "$OS" = "Windows_NT" ]; then
    echo "Error: Windows is not supported yet." 1>&2
    exit 1
  else
    UNAME_SM=$(uname -sm)
    case "$UNAME_SM" in
    "Darwin x86_64") target="darwin_amd64.tar.gz" ;;
    "Darwin arm64") target="darwin_arm64.tar.gz" ;;
    "Linux x86_64") target="linux_amd64.tar.gz" ;;
    "Linux aarch64") target="linux_arm64.tar.gz" ;;
    *) echo "Error: '$UNAME_SM' is not supported yet." 1>&2; exit 1 ;;
    esac
  fi

  # Check if plugin is provided as an argument
  if [ $# -eq 0 ] || [ -z "$1" ]; then
    printf "Enter the plugin name: "
    read plugin
  else
    plugin=$1
  fi

  # Check if version is provided as an argument
  if [ $# -lt 2 ] || [ -z "$2" ]; then
    printf "Enter the version (latest): "
    read version
    version=${version:-latest}
  else
    version=$2
  fi

  # Check if location is provided as an argument
  if [ $# -lt 3 ] || [ -z "$3" ]; then
    printf "Enter location (/usr/local/bin): "
    read location
    location=${location:-/usr/local/bin}
  else
    location=$3
  fi

  bin_dir=$location
  exe="$bin_dir/steampipe_export_${plugin}"

  tmp_dir=$(mktemp -d)
  mkdir -p "${tmp_dir}"
  tmp_dir="${tmp_dir%/}"

  echo "Created temporary directory at $tmp_dir."
  cd "$tmp_dir" || exit

  # set a trap for a clean exit - even in failures
  trap 'rm -rf $tmp_dir' EXIT

  case $(uname -s) in
    "Darwin" | "Linux") zip_location="$tmp_dir/steampipe_export_${plugin}.${target}" ;;
    *) echo "Error: steampipe_export_${plugin} is not supported on '$(uname -s)' yet." 1>&2; exit 1 ;;
  esac

  asset_name="steampipe_export_${plugin}.${target}"
  # Generate the URI for the binary
  if [ "$version" = "latest" ]; then
    uri="https://github.com/turbot/steampipe-plugin-${plugin}/releases/latest/download/${asset_name}"
  else
    uri="https://github.com/turbot/steampipe-plugin-${plugin}/releases/download/${version}/${asset_name}"
  fi

  echo ""
  echo "Downloading ${BOLD}${asset_name}${NORMAL}..."
  if ! curl --fail --location --progress-bar --output ${asset_name} "$uri"; then
    echo "Could not find version $version"
    exit 1
  fi

  echo "Deflating downloaded archive"
  tar -xvf "$zip_location" -C "$tmp_dir"

  echo "Installing"
  install -d "$bin_dir"
  install "$tmp_dir/steampipe_export_${plugin}" "$bin_dir"

  echo "Applying necessary permissions"
  chmod +x $exe

  echo "Removing downloaded archive"
  rm "$zip_location"

  echo "steampipe_export_${plugin} was installed successfully to $bin_dir"

  if ! command -v $bin_dir/steampipe_export_${plugin} >/dev/null 2>&1; then
    echo "steampipe_export_${plugin} was installed, but could not be executed. Are you sure '$bin_dir/steampipe_export_${plugin}' has the necessary permissions?"
    exit 1
  fi
}

# Call the main function to run the script
main "$@"
