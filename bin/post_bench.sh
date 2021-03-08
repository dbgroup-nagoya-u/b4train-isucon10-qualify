#!/bin/bash
set -ue -o pipefail

# global constants
readonly WORKSPACE=$(cd $(dirname ${BASH_SOURCE:-${0}})/../; pwd)
source ${WORKSPACE}/conf/env.sh

# run scripts at workspace
cd ${WORKSPACE}

####################################################################################################
# Documents
####################################################################################################

usage() {
  cat 1>&2 << EOS
Usage:
  ${BASH_SOURCE:-${0}} <branch_name>
Description:
  Perform bottleneck analysis and push its results to GitHub.
Arguments:
  <branch_name>: A branch name to push analysis results.
EOS
  exit 1
}

####################################################################################################
# Check command-line arguments
####################################################################################################

# check input arguments
if [ ${#} -ne 1 ]; then
  usage
fi

# check whether there is a specified branch
readonly GIT_BRANCH=${1}
git fetch origin
if ! git branch --list "${GIT_BRANCH}" | grep "${GIT_BRANCH}" &> /dev/null; then
  echo "There is no branch: ${GIT_BRANCH}" 1>&2
  exit 1
fi

####################################################################################################
# Bottleneck analysis
####################################################################################################

# clear old analysis results
rf -rf ${WORKSPACE}/bottleneck_analysis/*

# create UDF to analyze DB bottleneck
psql -h ${PG_HOST} -p ${PG_PORT} -U ${PG_USER} -d ${PG_DBNAME} \
  -f "${WORKSPACE}/conf/bottleneck_analysis/udf_analyze_slow_queries.sql"
# analyze slow queries
psql -h ${PG_HOST} -p ${PG_PORT} -U ${PG_USER} -d ${PG_DBNAME} \
  -f "${WORKSPACE}/conf/bottleneck_analysis/analyze_queries.sql" \
  > ${WORKSPACE}/bottleneck_analysis/db_summary.txt

# analyze app
ssh ${WEB_HOST} cat ${NGINX_LOG_PATH} \
  | kataribe -f "${WORKSPACE}/conf/bottleneck_analysis/kataribe.toml" \
  > ${WORKSPACE}/bottleneck_analysis/nginx_summary.txt

# push analysis results
git add ${WORKSPACE}/bottleneck_analysis
git commit -m "add bottleneck analysis results"
git push origin ${GIT_BRANCH}
