#!/bin/bash

# Copyright 2022 Authors of spidernet-io
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

PROJECT_ROOT=$(dirname ${BASH_SOURCE[0]})/../..

CHART_DIR=${1:-"${PROJECT_ROOT}/charts"}
# ====modify====
API_CODE_DIR=${2:-"${PROJECT_ROOT}/pkg/k8s/apis/rocktemplate.spidernet.io/v1"}

#======================

# CONST
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${PROJECT_ROOT}; ls -d -1 ./vendor/sigs.k8s.io/controller-tools/cmd/controller-gen 2>/dev/null || echo ../controller-gen)}

controllerGenCmd() {
  go run ${PROJECT_ROOT}/${CODEGEN_PKG}/main.go "$@"
}

echo "generate role yaml to chart"
controllerGenCmd rbac:roleName="exampleClusterRole" paths="${API_CODE_DIR}" output:stdout \
    | sed 's?name: exampleClusterRole?name: {{ include "project.name" . }}?' > ${CHART_DIR}/templates/role.yaml

echo "generate CRD yaml to chart"
controllerGenCmd crd paths="${API_CODE_DIR}"  output:crd:artifacts:config="${CHART_DIR}/crds"

echo "generate deepcode to api code"
controllerGenCmd crd paths="${API_CODE_DIR}"  object
