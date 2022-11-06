// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0

package mybookManager

import (
	"github.com/spidernet-io/rocktemplate/pkg/mybookManager/types"
	"go.uber.org/zap"
)

type mybookManager struct {
	logger     *zap.Logger
	webhook    *webhookhander
	controller *myController
}

var _ types.MybookManager = (*mybookManager)(nil)

func New(logger *zap.Logger) types.MybookManager {

	return &mybookManager{
		logger: logger.Named("mybookManager"),
	}
}
