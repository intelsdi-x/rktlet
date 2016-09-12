#!/bin/bash

# Copyright 2016 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This script exists to run the mockery tool and include licensing information
# in the output.

RKTLET_ROOT=$(dirname "${BASH_SOURCE}")/../..

set -o errexit
set -o nounset
set -o pipefail

function rktlet::scripts::mockery {
  local path=$1
  local interface=$2
  local destination=$3

  local year="$(date '+%Y')"
  local TMPFILE="/tmp/$(date +%s).go"
  echo "Running mockery for path ${path}, interface ${interface}"

  sed "s/YEAR/${year}/" "${RKTLET_ROOT}/hack/boilerplate/boilerplate.go.txt" > "$TMPFILE"
  cat >> "$TMPFILE" <<EOF
// Code generated by Mockery for ${interface}. This code should not be edited by hand
EOF

  "${RKTLET_ROOT}/hack/bin/mockery" -print -dir="${path}" -name="${interface}" >> "$TMPFILE"

  mv "$TMPFILE" "${destination}"
}

rktlet::scripts::mockery $@