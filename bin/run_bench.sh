#!/bin/bash
set -ue -o pipefail

# set global constants
source ${HOME}/env.sh

# run scripts at workspace
cd ${WORKSPACE}

####################################################################################################
# Documents
####################################################################################################

usage() {
  cat 1>&2 << EOS
Usage:
  ${BASH_SOURCE:-${0}} <target_host>
Description:
  Run benchmark.
Arguments:
  <target_host>: A host name of a benchmark target.
EOS
  exit 1
}

####################################################################################################
# Check command-line arguments
####################################################################################################

# check whether there is input arguments
if [ ${#} -ne 1 ]; then
  usage
fi
readonly TARGET_HOST=${1}

####################################################################################################
# Run benchmark
####################################################################################################

ssh bench "cd ${HOME}/isuumo/bench/ && ./bench --target-url http://${TARGET_HOST}"
