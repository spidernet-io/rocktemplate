// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0

package mybookManager

import (
	"context"
	"fmt"
	"github.com/spidernet-io/rocktemplate/pkg/k8s"
	crd "github.com/spidernet-io/rocktemplate/pkg/k8s/apis/rocktemplate.spidernet.io/v1"
	crdclientset "github.com/spidernet-io/rocktemplate/pkg/k8s/client/clientset/versioned"
	"github.com/spidernet-io/rocktemplate/pkg/k8s/client/informers/externalversions"
	crdlisterv1 "github.com/spidernet-io/rocktemplate/pkg/k8s/client/listers/rocktemplate.spidernet.io/v1"
	"github.com/spidernet-io/rocktemplate/pkg/lease"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"reflect"
	"time"
)

const (
	workerNumber    = 2
	queueMaxRetries = 100
)

type informerHandler struct {
	logger         *zap.Logger
	leaseName      string
	leaseNameSpace string
	leaseId        string
	eventRecord    record.EventRecorder
	queue          workqueue.RateLimitingInterface
	crdlister      crdlisterv1.MybookLister
	k8sclient      crdclientset.Interface
}

func (s *informerHandler) worker(ctx context.Context) {
	for s.processNextWorkItem(ctx) {
	}
}

func (s *informerHandler) processNextWorkItem(ctx context.Context) bool {
	key, quit := s.queue.Get()
	if quit {
		return false
	}
	defer s.queue.Done(key)

	err := s.syncHandler(ctx, key.(*crd.Mybook))
	if err == nil {
		s.queue.Forget(key)
	} else {
		s.logger.Sugar().Warnf("worker failed , error=%v", err)
		if apierrors.IsConflict(err) {
			// 更新CRD 时，resourceVersion 冲突，重试
			s.queue.AddRateLimited(key)
		} else if s.queue.NumRequeues(key) < queueMaxRetries {
			s.queue.AddRateLimited(key)
		}
	}
	// handle nex item
	return true
}

func (s *informerHandler) syncHandler(ctx context.Context, obj *crd.Mybook) error {
	if obj == nil {
		return nil
	}
	logger := s.logger.Named("worker")

	// 通过 clientset 向 api server 实时获取最新数据
	// old, err := s.k8sclient.RocktemplateV1().Mybooks().Get(ctx, obj.Name, metav1.GetOptions{})
	// 获取最新cache中的数据（cache中的数据有延时风险）
	old, err := s.crdlister.Get(obj.Name)
	if err != nil || old == nil {
		logger.Warn("failed to get " + obj.Name)
		return nil
	}
	logger.Info("handle " + obj.Name)

	newone := old.DeepCopy()
	newone.Status.TotalIPCount = 100

	if !reflect.DeepEqual(old, newone) {
		if _, err := s.k8sclient.RocktemplateV1().Mybooks().UpdateStatus(ctx, newone, metav1.UpdateOptions{}); err != nil {
			// if conflicted, queue will retry it later
			return err
		}
	}

	return nil
}

// ===================================

func (s *informerHandler) informerAddHandler(obj interface{}) {
	s.logger.Sugar().Infof("start crd add: %+v", obj)

	r, ok := obj.(*crd.Mybook)
	if !ok {
		s.logger.Sugar().Errorf("failed to get crd: %+v", obj)
		return
	}
	s.logger.Sugar().Infof("mybook crd: %+v", r)

	// enqueue
	s.queue.AddRateLimited(r)

	// generate crd event
	s.eventRecord.Eventf(r, corev1.EventTypeNormal, "newMybook", "crd event, new mybook %v", r.Name)

	s.logger.Sugar().Infof("done crd add: %+v", obj)
}

func (s *informerHandler) informerUpdateHandler(oldObj interface{}, newObj interface{}) {
	s.logger.Sugar().Infof("crd update old: %+v", oldObj)
	s.logger.Sugar().Infof("crd update new: %+v", newObj)

}

func (s *informerHandler) informerDeleteHandler(obj interface{}) {
	s.logger.Sugar().Infof("crd delete: %+v", obj)
}

// ===================================

func (s *informerHandler) executeInformer() {

	// ------- client set
	config, err := rest.InClusterConfig()
	if err != nil {
		s.logger.Sugar().Fatalf("failed to InClusterConfig, reason=%v", err)
	}
	clientset, err := crdclientset.NewForConfig(config) // 初始化 client
	if err != nil {
		s.logger.Sugar().Fatalf("failed to NewForConfig, reason=%v", err)
	}
	s.k8sclient = clientset

	ctx, cancel := context.WithCancel(context.Background())

	if len(s.leaseName) > 0 && len(s.leaseNameSpace) > 0 && len(s.leaseId) > 0 {
		s.logger.Sugar().Infof("%v try to get lease %s/%s to run informer", s.leaseId, s.leaseNameSpace, s.leaseName)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		rlogger := s.logger.Named(fmt.Sprintf("lease %s/%s", s.leaseNameSpace, s.leaseName))
		// id := globalConfig.PodName
		getLease, lossLease, err := lease.NewLeaseElector(ctx, s.leaseNameSpace, s.leaseName, s.leaseId, rlogger)
		if err != nil {
			s.logger.Sugar().Fatalf("failed to generate lease, reason=%v ", err)
		}
		<-getLease
		s.logger.Sugar().Infof("succeed to get lease %s/%s to run informer", s.leaseNameSpace, s.leaseName)

		go func(lossLease chan struct{}) {
			<-lossLease
			cancel()
			s.logger.Sugar().Warnf("lease %s/%s is loss, informer is broken", s.leaseNameSpace, s.leaseName)
		}(lossLease)
	}

	// setup informer
	s.logger.Info("begin to setup informer")
	factory := externalversions.NewSharedInformerFactory(clientset, 0)
	// 注意，一个 factory 下  对同一种 CRD 不能 创建 多个Informer，不然会 数据竞争 问题。 而 一个 factory 下， 可对不同 CRD 产生 各种的 Informer

	t := factory.Rocktemplate().V1().Mybooks()
	s.crdlister = t.Lister()

	inform := t.Informer()

	// 在一个 Handler 逻辑中，是顺序消费所有的 crd 事件的
	// 简单说：有2个 crd add 事件，那么，先会调用 informerAddHandler 完成 事件1 后，才会 调用 informerAddHandler 处理 事件2
	inform.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    s.informerAddHandler,
		UpdateFunc: s.informerUpdateHandler,
		DeleteFunc: s.informerDeleteHandler,
	})

	// 一个 inform 下  如果注册 第二套 AddEventHandler，那么，对于同一个 事件，两套 handler 是 使用 独立协程 并发调用的 .
	// 这样，就能实现对同一个事件 并发调用不同的回调，好处是，他们底层是基于同一个 NewSharedInformer ， 共用一个cache，能降低api server 之间的数据同步
	inform.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    s.informerAddHandler,
		UpdateFunc: s.informerUpdateHandler,
		DeleteFunc: s.informerDeleteHandler,
	})

	defer s.queue.ShutDown()
	go func() {
		s.logger.Info("start worker")

		if !cache.WaitForNamedCacheSync("deployment", ctx.Done(), inform.HasSynced) {
			s.logger.Error("failed to sync cache")
			cancel()
			return
		}

		for i := 0; i < workerNumber; i++ {
			go wait.UntilWithContext(ctx, s.worker, time.Second)
		}
	}()

	inform.Run(ctx.Done())

}

func (s *mybookManager) RunInformer(leaseName, leaseNameSpace string, leaseId string) {

	scheme, e := crd.SchemeBuilder.Build()
	if e != nil {
		s.logger.Sugar().Fatalf("failed to get crd scheme: %+v", e)
	}
	/*
		Events:
		  Type    Reason     Age   From    Message
		  ----    ------     ----  ----    -------
		  Normal  newMybook  13s   mybook  crd event, new mybook test
	*/
	p := k8s.NewEventRecord(scheme, "mybook", s.logger)

	// -----------
	t := &informerHandler{
		logger:         s.logger,
		leaseName:      leaseName,
		leaseNameSpace: leaseNameSpace,
		leaseId:        leaseId,
		eventRecord:    p,
		queue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "mybook"),
	}
	s.informer = t

	go func() {
		for {
			t.executeInformer()
			time.Sleep(time.Duration(5) * time.Second)
		}
	}()
}
