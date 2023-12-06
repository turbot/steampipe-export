#!/usr/bin/env bash

set -e

main() {
  # ANSI escape code variables
  BOLD=$(tput bold)
  NORMAL=$(tput sgr0)

  if ! command -v tar >/dev/null; then
    echo "Error: 'tar' is required." 1>&2
    exit 1
  fi

  if [ "$OS" = "Windows_NT" ]; then
    echo "Error: Windows is not supported yet." 1>&2
    exit 1
  else
    case $(uname -sm) in
    "Darwin x86_64") target="darwin_amd64.tar.gz" ;;
    "Darwin arm64") target="darwin_arm64.tar.gz" ;;
    "Linux x86_64") target="linux_amd64.tar.gz" ;;
    "Linux aarch64") target="linux_arm64.tar.gz" ;;
    *) echo "Error: '$(uname -sm)' is not supported yet." 1>&2;exit 1 ;;
    esac
  fi

  # Check if plugin is provided as an argument
  if [ $# -eq 0 ] || [ -z "$1" ]; then
    read -p "Enter the plugin name: " plugin
  else
    plugin=$1
  fi

  # Check if version is provided as an argument
  if [ $# -lt 2 ] || [ -z "$2" ]; then
    read -p "Enter the version (default: latest): " version
    version=${version:-latest}  # Default to 'latest' if input is empty
  else
    version=$2
  fi

  bin_dir="/usr/local/bin"
  exe="$bin_dir/steampipe_export_${plugin}"

  test -z "$tmp_dir" && tmp_dir="$(mktemp -d)"
  mkdir -p "${tmp_dir}"
  tmp_dir="${tmp_dir%/}"

  echo "Created temporary directory at $tmp_dir."
  cd "$tmp_dir"

  # set a trap for a clean exit - even in failures
  trap 'rm -rf $tmp_dir' EXIT

  case $(uname -s) in
    "Darwin") zip_location="$tmp_dir/steampipe_export_${plugin}.${target}" ;;
    "Linux") zip_location="$tmp_dir/steampipe_export_${plugin}.${target}" ;;
    *) echo "Error: steampipe_export_${plugin} is not supported on '$(uname -s)' yet." 1>&2;exit 1 ;;
  esac

  # Generate the URI for the binary
  if [ "$version" = "latest" ]; then
    uri="https://api.github.com/repos/turbotio/steampipe-plugin-${plugin}/releases/latest"
    asset_name="steampipe_export_${plugin}.${target}"
  else
    uri="https://api.github.com/repos/turbotio/steampipe-plugin-${plugin}/releases/tags/${version}"
    asset_name="steampipe_export_${plugin}.${target}"
  fi

  # Read the GitHub Personal Access Token
  GITHUB_TOKEN=${GITHUB_TOKEN:-}  # Assuming GITHUB_TOKEN is set as an environment variable

  # Check if the GITHUB_TOKEN is set
  if [ -z "$GITHUB_TOKEN" ]; then
    echo ""
    echo "Error: GITHUB_TOKEN is not set. Please set your GitHub Personal Access Token as an environment variable." 1>&2
    exit 1
  fi
  AUTH="Authorization: token $GITHUB_TOKEN"

  response=$(curl -sH "$AUTH" $uri)
  id=`echo "$response" | jq --arg asset_name "$asset_name" '.assets[] | select(.name == $asset_name) | .id' |  tr -d '"'`
  GH_ASSET="$uri/releases/assets/$id"

  echo ""
  echo "Downloading ${BOLD}${asset_name}${NORMAL}..."
  curl -#SL -H "$AUTH" -H "Accept: application/octet-stream" \
     "https://api.github.com/repos/turbotio/steampipe-plugin-${plugin}/releases/assets/$id" \
     -o "$asset_name" -L --create-dirs --output "$zip_location"

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

  if ! command -v $bin_dir/steampipe_export_${plugin}  >/dev/null; then
    echo "steampipe_export_${plugin} was installed, but could not be executed. Are you sure '$bin_dir/steampipe_export_${plugin}' has the necessary permissions?"
    exit 1
  fi

}

# Call the main function to run the script
main "$@"
