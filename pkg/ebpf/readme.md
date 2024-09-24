
(1)  k8s 集群内实现 service 解析，包括 clusterip 、 nodePort 等 

（2）支持集群外 主机部署， 实现 主机应用 直接访问到 pod ip（macvlan）或者  pod 所在主机 ip + nodePort 
	这样，能够避免传统 nodePort 等方案带来的 源端口冲突、并发低、转发性能差 等问题 

    尤其是 kubevirt 虚拟机场景

(3) 支持 localRedirect，支持 local dns 

（4） 为多集群 如 K3S 而服务

（5）支持 kubeedge， 在 边端 不需要 cni 的情况下，边端进行 clusterIP 解析，把流量 封发到 云端 
     pod 所在节点的 nodePort
