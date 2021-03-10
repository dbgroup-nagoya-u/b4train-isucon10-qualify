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
git fetch --quiet origin
if ! git branch --list "${GIT_BRANCH}" | grep "${GIT_BRANCH}" &> /dev/null; then
  echo "There is no branch: ${GIT_BRANCH}" 1>&2
  exit 1
fi

####################################################################################################
# Bottleneck analysis
####################################################################################################

# clear old analysis results
echo "remove old analysis results"
rm -rf ${WORKSPACE}/bottleneck_analysis/*

# create UDF to analyze DB bottleneck
echo "analyze DB bottleneck"
psql -h ${PG_HOST} -p ${PG_PORT} -U ${PG_USER} -d ${PG_DBNAME} \
  -f "${WORKSPACE}/conf/bottleneck_analysis/udf_analyze_slow_queries.sql"
# analyze slow queries
psql -h ${PG_HOST} -p ${PG_PORT} -U ${PG_USER} -d ${PG_DBNAME} \
  -f "${WORKSPACE}/conf/bottleneck_analysis/analyze_queries.sql" \
  > ${WORKSPACE}/bottleneck_analysis/db_summary.txt

# analyze app
echo "analyze App bottleneck"
ssh ${WEB_HOST} cat ${NGINX_LOG_PATH} \
  | kataribe -f "${WORKSPACE}/conf/bottleneck_analysis/kataribe.toml" \
  > ${WORKSPACE}/bottleneck_analysis/nginx_summary.txt

# push analysis results
echo "push analysis results to ${GIT_BRANCH}"
git add ${WORKSPACE}/bottleneck_analysis
git commit --quiet -m "add bottleneck analysis results"
git push --quiet origin ${GIT_BRANCH}

echo "done."
