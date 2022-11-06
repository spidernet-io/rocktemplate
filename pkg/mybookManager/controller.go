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
	lister "github.com/spidernet-io/rocktemplate/pkg/k8s/client/listers/rocktemplate.spidernet.io/v1"
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
	crdKindName = "Mybook"

	// maxRetries is the number of times it will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(maxRetries-1)) the following numbers represent the times
	// 5ms, 10ms, 20ms, 40ms, 80ms, 160ms, 320ms, 640ms, 1.3s, 2.6s, 5.1s, 10.2s, 20.4s, 41s, 82s
	maxRetries = 15
)

type myController struct {
	logger         *zap.Logger
	leaseName      string
	leaseNameSpace string
	leaseId        string
	// generate k8s event
	eventRecord       record.EventRecorder
	clientset         *crdclientset.Clientset
	crdLister         lister.MybookLister
	queueEvent        workqueue.RateLimitingInterface
	eventWorkerNumber int
}

func (s *myController) informerAddHandler(obj interface{}) {

	r, ok := obj.(*crd.Mybook)
	if !ok {
		s.logger.Sugar().Errorf("failed to get crd: %v", r.Name)
		return
	}
	s.logger.Sugar().Infof("informer add crd: %+v", r.Name)

	// time.Sleep(30 * time.Second)
	s.logger.Sugar().Infof("done crd add: %+v", r.Name)
}

func (s *myController) informerUpdateHandler(oldObj interface{}, newObj interface{}) {
	curPod := newObj.(*crd.Mybook)
	oldPod := oldObj.(*crd.Mybook)
	if curPod.ResourceVersion == oldPod.ResourceVersion {
		// Periodic resync will send update events for all known pods.
		// Two different versions of the same pod will always have different RVs.
		return
	}
	s.logger.Sugar().Infof("informer update crd: %+v", curPod.Name)

	// // 简单方式处理事件，堵塞执行，重建重试5次更新，
	// // 好处是代码简单，坏处是，resourceVersion、断网等场景，可能最终5次失败而最终失败
	// if !reflect.DeepEqual(curPod.Spec, oldPod.Spec) {
	// 	// when errors.IsConflict owing to resourceVersion, auto retry 5 times at interval 10ms
	// 	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
	// 		return s.handleCrdEvent(context.Background(), curPod)
	// 	})
	// 	if err != nil {
	// 		s.logger.Sugar().Errorf("failed to update mybook status, error=%v", err)
	// 	}
	// }

	// 基于工作队列，生产者和消费者模型，使得事件能够本可靠的 异步执行 成功
	// 例如 resourceVersion、断网等失败，最终都能够不断重试而手工
	if !reflect.DeepEqual(curPod.Spec, oldPod.Spec) {
		// AddRateLimited adds an item to the workqueue after the rate limiter says it's ok
		s.queueEvent.AddRateLimited(curPod)
	}

}

func (s *myController) informerDeleteHandler(obj interface{}) {
	curPod := obj.(*crd.Mybook)
	s.logger.Sugar().Infof("informer delete crd: %v", curPod.Name)
}

// --------------------

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
// It enforces that the syncHandler is never invoked concurrently with the same key.
func (s *myController) crdEventWorker(ctx context.Context) {
	// handle all queued item and exit
	for s.processNextqueueEventItem(ctx) {
	}
}

func (s *myController) processNextqueueEventItem(ctx context.Context) bool {
	key, quit := s.queueEvent.Get()
	if quit {
		return false
	}
	defer s.queueEvent.Done(key)

	r := key.(*crd.Mybook)
	s.logger.Sugar().Debugf("process item %v ， current queue length %v", r.Name, s.queueEvent.Len())
	err := s.handleCrdEvent(ctx, r)
	if err == nil {
		// succeed to handle the event
		s.queueEvent.Forget(key)
	} else {
		if s.queueEvent.NumRequeues(key) < maxRetries {
			// failed to handle the event, add back to the queue and retry later
			s.queueEvent.AddRateLimited(key)
		} else {
			s.logger.Sugar().Errorf("dropping event %+v", key)
			s.queueEvent.Forget(key)
		}
	}

	// process next queue item
	return true
}

// 添加controler业务
func (s *myController) handleCrdEvent(ctx context.Context, obj *crd.Mybook) error {
	s.logger.Sugar().Infof("handle Crd Event: %v", obj.Name)

	// 从 informer 缓存中获取数据，可能因为延时 而不是最新
	t, e := s.crdLister.Get(obj.Name)
	// 从 api server 获取最实时的数据
	// t, e := s.clientset.RocktemplateV1().Mybooks().Get(ctx, obj.Name, metav1.GetOptions{})
	if e != nil {
		if apierrors.IsNotFound(e) {
			// not found ,no retry
			return nil
		}
		// retry later
		return e
	}

	t.Status.TotalIPCount = 100

	if _, e := s.clientset.RocktemplateV1().Mybooks().UpdateStatus(ctx, t, metav1.UpdateOptions{}); e != nil {
		if apierrors.IsConflict(e) {
			// resrouceVersion Conflict, retry later
			return e
		}
		return e
	} else {
		s.logger.Sugar().Infof("succeed to update mybook status ")
	}

	// generate crd event
	s.eventRecord.Eventf(t, corev1.EventTypeNormal, "modified Mybook", "crd event, new mybook %v", t.Name)

	return nil
}

// ===================================

func (s *myController) executeInformerOnce() {
	defer func() {
		s.logger.Warn("controller down")
	}()
	s.logger.Info("controller up")

	// ------- client set
	config, err := rest.InClusterConfig()
	if err != nil {
		s.logger.Sugar().Fatalf("failed to InClusterConfig, reason=%v", err)
	}
	clientset, err := crdclientset.NewForConfig(config) // 初始化 client
	if err != nil {
		s.logger.Sugar().Fatalf("failed to NewForConfig, reason=%v", err)
	}
	s.clientset = clientset

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// --------- 是否 基于 lease 来 选主，从而 启动 informer
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
	factory := externalversions.NewSharedInformerFactory(clientset, 0)
	// 注意，一个 factory 下  对同一种 CRD 不能 创建 多个Informer，不然会 数据竞争 问题。 而 一个 factory 下， 可对不同 CRD 产生 各种的 Informer
	inform := factory.Rocktemplate().V1().Mybooks().Informer()

	// 在一个 Handler 逻辑中，是顺序消费所有的 crd 事件的
	// 简单说：有2个 crd add 事件，那么，先会调用 informerAddHandler 完成 事件1 后，才会 调用 informerAddHandler 处理 事件2
	inform.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    s.informerAddHandler,
		UpdateFunc: s.informerUpdateHandler,
		DeleteFunc: s.informerDeleteHandler,
	})
	s.crdLister = factory.Rocktemplate().V1().Mybooks().Lister()

	// // 一个 inform 下  如果注册 第二套 AddEventHandler，那么，对于同一个 事件，两套 handler 是 使用 独立协程 并发调用的 .
	// // 这样，就能实现对同一个事件 并发调用不同的回调，好处是，他们底层是基于同一个 NewSharedInformer ， 共用一个cache，能降低api server 之间的数据同步
	// inform.AddEventHandler(cache.ResourceEventHandlerFuncs{
	// 	AddFunc:    s.informerAddHandler,
	// 	UpdateFunc: s.informerUpdateHandler,
	// 	DeleteFunc: s.informerDeleteHandler,
	// })

	go func() {
		s.logger.Debug("informer up")
		inform.Run(ctx.Done())
		s.logger.Debug("informer down")
	}()

	// run event handler
	if s.eventWorkerNumber > 0 {
		go func() {
			defer s.queueEvent.ShutDown()

			// worker 中 如果使用了 lister 来获取缓存数据，此处需要等待 数据同步完毕
			s.logger.Sugar().Debugf("wait for Cache Sync")
			if !cache.WaitForNamedCacheSync(crdKindName, ctx.Done(), inform.HasSynced) {
				return
			}

			s.logger.Sugar().Debugf("start worker with counts %v ", s.eventWorkerNumber)
			for i := 0; i < s.eventWorkerNumber; i++ {
				go wait.UntilWithContext(ctx, s.crdEventWorker, time.Second)
			}
			<-ctx.Done()
		}()
	}

	<-ctx.Done()

}

func (s *mybookManager) RunController(leaseName, leaseNameSpace string, leaseId string) {

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
	p := k8s.NewEventRecord(scheme, crdKindName, s.logger)

	// -----------
	t := &myController{
		logger:            s.logger,
		leaseName:         leaseName,
		leaseNameSpace:    leaseNameSpace,
		leaseId:           leaseId,
		eventRecord:       p,
		queueEvent:        workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), crdKindName),
		eventWorkerNumber: 2,
	}
	s.controller = t

	go func() {
		for {
			t.executeInformerOnce()
			time.Sleep(time.Duration(5) * time.Second)
		}
	}()
}
