#!/bin/bash

# Copyright 2022 Authors of spidernet-io
# SPDX-License-Identifier: Apache-2.0

# generate skd for the crd of client,informer,lister, to pkg/k8s/client

set -o errexit
set -o nounset
set -o pipefail

# refer to https://github.com/kubernetes/sample-controller/blob/master/hack/update-codegen.sh

APIS_PKG="pkg/k8s/apis"
OUTPUT_PKG="pkg/k8s/client"

#===================

PROJECT_ROOT=$(git rev-parse --show-toplevel)
CODEGEN_PKG=${CODEGEN_PKG_PATH:-$(cd ${PROJECT_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}
MODULE_NAME=$(cat ${PROJECT_ROOT}/go.mod | grep -e "module[[:space:]][^[:space:]]*" | awk '{print $2}')

SPDX_COPYRIGHT_HEADER="${PROJECT_ROOT}/tools/copyright-header.txt"

TMP_DIR="${PROJECT_ROOT}/output/codeGen"
LICENSE_FILE="${TMP_DIR}/boilerplate.go.txt"
GO_PATH_DIR="${TMP_DIR}/go"

rm -rf ${TMP_DIR}
mkdir -p ${TMP_DIR}

touch ${LICENSE_FILE}
while read -r line || [[ -n ${line} ]]
do
    echo "// ${line}" >>${LICENSE_FILE}
done < ${SPDX_COPYRIGHT_HEADER}

cd "${PROJECT_ROOT}"

rm -rf ${OUTPUT_PKG} || true

# https://github.com/kubernetes/code-generator/blob/master/kube_codegen.sh
source "${PROJECT_ROOT}/${CODEGEN_PKG}/kube_codegen.sh"

kube::codegen::gen_helpers \
    --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.go.txt" \
    "${SCRIPT_ROOT}/pkg/apis"

kube::codegen::gen_client\
    --with-watch \
    --output-dir "${PROJECT_ROOT}/${OUTPUT_PKG}" \
    --output-pkg "${MODULE_NAME}/${OUTPUT_PKG}" \
    --boilerplate ${LICENSE_FILE} \
    "${PROJECT_ROOT}/${APIS_PKG}"

rm -rf ${TMP_DIR}
exit 0

