#ifndef __MAP_H__
#define __MAP_H__


#include "vmlinux.h"


//======================================= map ： 存储 service ， 包括  service( clusterip nodePort  loadbalancer ) localRedirect  loadbalancer

#define DEFAULT_MAX_EBPF_MAP_ENTRIES 65536

// map 的 key 和 value struct，会自动在 golang 中 生成 struct

#ifdef ENABLE_IPV4

#define NAT_TYPE_SERVICE    0
#define NAT_TYPE_REDIRECT   1
#define NAT_TYPE_BALANCING    2

#define NODE_PORT_IP	0xffffffff

/*
    对于 NAT_TYPE_SERVICE ：
        无论目的是 clusterIP，nodePort，LoadBalancer  ， 都是解析到 pod ip
    对于 NAT_TYPE_REDIRECT：
        解析到 本地 pod ip，
*/
// a loadbalancer has several entries mapping to each backend
struct mapkey_service {
  ipv4_addr_t address;     /* 小端存储。 clusterIP， 或者 NODE_PORT_IP(255.255.255.255) 表示 nodePort  */
  __be16 dport;            /* 小端存储。 clusterIP 的 端口， 或者 nodePort 的 端口     */
  __u8  proto;
  __u8  nat_type;         /* NAT_TYPE_SERVICE (  lowest priority  ) ,NAT_TYPE_REDIRECT ,  NAT_TYPE_BALANCING ( highest priority )  */
  __u8  scope;
  __u8  pad[3];
};


#define SERVICE_FLAG_EXTERNAL_LOCAL_SVC	0x1
#define SERVICE_FLAG_INTERNAL_LOCAL_SVC	0x2

#define NAT_FLAG_ACCESS_NODEPORT_BALANCING	0x1

#define NAT_FLAG_ALLOW_ACCESS_SERVICE	0x1

struct mapvalue_service {
  __u32 svc_id ;                 // 一个 service 有一个 唯一的 ID ，用来映射 service 下 所有的 endpoint
  __u32 total_backend_count;         // how many global backend exist in the service
  __u32 local_backend_count;         // how many local-node backend exist in the service ，用于实现 clientIP 亲和
  __u32 affinity_second;       /* In seconds, only for svc frontend */
  __u8  service_flags;                /* SERVICE_FLAG_EXTERNAL_LOCAL_SVC  , SERVICE_FLAG_INTERNAL_LOCAL_SVC */
  __u8  balancing_flags;                /* NAT_FLAG_ACCESS_NODEPORT_BALANCING（是打到 pod 所在节点的 nodePort，还是 pod ip）  */
  __u8  redirect_flags;         /* NAT_FLAG_ALLOW_ACCESS_SERVICE( 如果在 local-node backend 不可用时，是否正常解析到 clusterIP)  */
  __u8  pad;
};


struct {
  __uint(type, BPF_MAP_TYPE_HASH);
  __type(key, struct mapkey_service  );
  __type(value, struct mapvalue_service  );
  __uint(pinning, 1); /* 这个配合 golang 中的 pinPath，完成 路径 pin */
  //__uint(map_flags, 0);
  __uint(max_entries, DEFAULT_MAX_EBPF_MAP_ENTRIES);
} map_service SEC(".maps");


//======================================= map ： 存储 endpoint  ， pod ip

struct mapkey_backend {
    __be32 order;      //  第几个 endpoint ip 。 前面几个记录，优先存储 本地 node 上的 endpoint ， 用于实现 clientIP 亲和
    __u32 svc_id;  // 对应 mapvalue_service 中的 svc_id ，  一个 service 有一个 唯一的 ID ，用来映射 service 下 所有的 endpoint
    __be16 dport;
    __u8  proto;
    __u8  nat_type;
    __u8  scope;
    __u8  pad[3];
};

struct mapvalue_backend {
	ipv4_addr_t pod_address;		/* 小端存储。 Service endpoint IPv4 address , saved in LittleEndian */
	ipv4_addr_t node_address;		/* 小端存储。 for loadbalancer , access the nodePort */
	__be16 pod_port;		/* 小端存储。 L4 port filter , saved in LittleEndian */
	__be16 node_port;		/* 小端存储。 for loadbalancer , access the nodePort */
};

struct {
  __uint(type, BPF_MAP_TYPE_HASH);
  __type(key, struct mapkey_backend );
  __type(value, struct mapvalue_backend );
  __uint(max_entries, DEFAULT_MAX_EBPF_MAP_ENTRIES);
  __uint(pinning, 1); /* 这个配合 golang 中的 pinPath，完成 路径 pin */
  //__uint(map_flags, 0);
} map_backend SEC(".maps");


//-----------------------  map ： 存储 node ip ，用于匹配 nodePort 中的 主机 ip


struct mapkey_node {
	__u32 address;       // ip
};

struct {
  __uint(type, BPF_MAP_TYPE_HASH );
  __type(key, struct mapkey_node  );
  __type(value, __u32  );
  __uint(max_entries, DEFAULT_MAX_EBPF_MAP_ENTRIES);
  __uint(pinning, 1); /* 这个配合 golang 中的 pinPath，完成 路径 pin */
  //__uint(map_flags, 0);
} map_node SEC(".maps");


//======================================= map ： 存储 亲和 记录

struct mapkey_affinity {
   __u64 client_cookie;       //  bpf_get_socket_cookie(ctx);
   ipv4_addr_t original_dest_ip;  // 小端存储。
   //ipv4_addr_t client_ip;
   __u16 original_port ;   // 小端存储。
   __u8 proto;
   __u8 pad;
};

struct mapvalue_affinity {
   __u64 ts;   // 这个值存储了 上次发生 亲和命中的 时间
   ipv4_addr_t nat_ip;   // 小端存储。
   __u16       nat_port ;   // 小端存储。
   __u8 proto;
   __u8 pad;
};


struct {
  __uint(type, BPF_MAP_TYPE_LRU_HASH);
  __type(key, struct mapkey_affinity  );
  __type(value, struct  mapvalue_affinity );
  __uint(max_entries, DEFAULT_MAX_EBPF_MAP_ENTRIES);
  __uint(pinning, 1); /* 这个配合 golang 中的 pinPath，完成 路径 pin */
  //__uint(map_flags, 0);
} map_affinity SEC(".maps");


//======================================= map ： 存储 nat 记录


struct mapkey_nat_record {
   __u64 socket_cookie;       //  bpf_get_socket_cookie(ctx);
   ipv4_addr_t nat_ip;    //   小端存储。  nat 后的 ip
   __u16 nat_port;       // 小端存储。
   __u8 proto;
   __u8 pad;
};

struct mapvalue_nat_record {
	ipv4_addr_t original_dest_ip;  // 小端存储。
	__u16 original_dest_port;       // 小端存储。
   __u8 pad[2];
};


struct {
  __uint(type, BPF_MAP_TYPE_LRU_HASH);
  __type(key, struct mapkey_nat_record );
  __type(value, struct mapvalue_nat_record );
  __uint(max_entries, DEFAULT_MAX_EBPF_MAP_ENTRIES);
  __uint(pinning, 1); /* 这个配合 golang 中的 pinPath，完成 路径 pin */
  //__uint(map_flags, 0);
} map_nat_record SEC(".maps");

#endif



//======================================= map ： 存储 event


// BPF_MAP_TYPE_PERF_EVENT_ARRAY 类型的 数据结构体，在 golang 中需要自定定义 struct

struct event_value {
    ipv4_addr_t nat_ip ;    // 小端存储。
    ipv4_addr_t original_dest_ip ;  /* 小端存储。 dest ip */
	__be16 nat_port;   // 小端存储。
	__be16 original_dest_port;   // 小端存储。
	__u32  tgid;
    __u8   is_ipv4 ; /* 0 for ipv6 data, 1 for ipv4 data */
    __u8   is_success ; /* 1 for success , 0 for failure */
    __u8   nat_type ;  /* NAT_TYPE_SERVICE (  lowest priority  ) ,NAT_TYPE_REDIRECT ,  NAT_TYPE_BALANCING ( highest priority )  */
    __u8   pad;
} ;

// BPF_MAP_TYPE_PERF_EVENT_ARRAY 中的 key 和 value 并不存放真正的 数据， key 用来存放 cpu 索引， values 存放指向 perf event buffer 的地址
struct {
	__uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY );
	__uint(key_size, sizeof(__u32));  /* BPF_MAP_TYPE_PERF_EVENT_ARRAY */
	__uint(value_size, sizeof(__u32));
	__uint(pinning, 1);  /* 这个配合 golang 中的 pinPath，完成 路径 pin */
	__uint(max_entries, 4096); /*默认会将max_entries设置为系统中cpu个数*/
} map_event  SEC(".maps");

// BPF_MAP_TYPE_RINGBUF
//struct {
//	__uint(type, BPF_MAP_TYPE_RINGBUF);
//	__uint(max_entries, 1 << 24);
//} map_event SEC(".maps");


#endif /* __MAP_H__ */
