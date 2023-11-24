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
    "Darwin x86_64") target="darwin_amd64.gz" ;;
    "Darwin arm64") target="darwin_arm64.gz" ;;
    "Linux x86_64") target="linux_amd64.gz" ;;
    "Linux aarch64") target="linux_arm64.gz" ;;
    *) echo "Error: '$(uname -sm)' is not supported yet." 1>&2;exit 1 ;;
    esac
  fi

  # Check if the correct number of arguments is given
  if [ $# -eq 0 ]; then
    echo "Usage: $0 <plugin> [version]"
    exit 1
  else
    plugin=$1
    version=${2:-latest}
  fi

  bin_dir="/usr/local/bin"
  exe="$bin_dir/${plugin}_dump"

  test -z "$tmp_dir" && tmp_dir="$(mktemp -d)"
  mkdir -p "${tmp_dir}"
  tmp_dir="${tmp_dir%/}"

  echo "Created temporary directory at $tmp_dir."
  cd "$tmp_dir"

  # set a trap for a clean exit - even in failures
  trap 'rm -rf $tmp_dir' EXIT

  case $(uname -s) in
    "Darwin") zip_location="$tmp_dir/${plugin}_dump_${target}" ;;
    "Linux") zip_location="$tmp_dir/${plugin}_dump_${target}" ;;
    *) echo "Error: ${plugin}_dump is not supported on '$(uname -s)' yet." 1>&2;exit 1 ;;
  esac

  # Generate the URI for the binary
  if [ "$version" = "latest" ]; then
    uri="https://api.github.com/repos/turbotio/steampipe-plugin-${plugin}/releases/latest"
    asset_name="${plugin}_dump_${target}"
  else
    uri="https://api.github.com/repos/turbotio/steampipe-plugin-${plugin}/releases/tags/${version}"
    asset_name="${plugin}_dump_${target}"
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
  tar -xf "$zip_location" -C "$tmp_dir"

  echo "Installing"
  install -d "$bin_dir"
  install "$tmp_dir/${plugin}_dump" "$bin_dir"

  echo "Applying necessary permissions"
  chmod +x $exe

  echo "Removing downloaded archive"
  rm "$zip_location"

  echo "${plugin}_dump was installed successfully to $bin_dir"

  if ! command -v $bin_dir/${plugin}_dump  >/dev/null; then
    echo "${plugin}_dump was installed, but could not be executed. Are you sure '$bin_dir/${plugin}_dump' has the necessary permissions?"
    exit 1
  fi

}

# Call the main function to run the script
main "$@"