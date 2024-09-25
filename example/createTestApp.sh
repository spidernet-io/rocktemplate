NAME=http-server
NAMESPACE=default
IMAGE=localhost/weizhoulan/rdmatool:latest

cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: ${NAME}
  namespace: ${NAMESPACE}
  labels:
    app: $NAME
spec:
  selector:
    matchLabels:
      app: $NAME
  template:
    metadata:
      name: $NAME
      labels:
        app: $NAME
    spec:
      containers:
      - name: $NAME
        image: $IMAGE
        imagePullPolicy: IfNotPresent
        command: ["/usr/bin/agent"]
        args: ["--app-mode=true"]
        securityContext:
          privileged: true
        ports:
        - containerPort: 80
          name: http
---
apiVersion: v1
kind: Service
metadata:
  name: $NAME-v4
  namespace: ${NAMESPACE}
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 80
    name: http1
  - port: 8080
    targetPort: 80
    name: http2
  selector:
    app: $NAME
---
apiVersion: v1
kind: Service
metadata:
  name: $NAME-affinity-v4
  namespace: ${NAMESPACE}
spec:
  type: NodePort
  sessionAffinity: ClientIP
  sessionAffinityConfig:
    clientIP:
      timeoutSeconds: 60
  ports:
  - port: 80
    targetPort: 80
    name: http
  selector:
    app: $NAME
---
apiVersion: v1
kind: Service
metadata:
  name: $NAME-local-v4
  namespace: ${NAMESPACE}
spec:
  type: NodePort
  externalTrafficPolicy: Local
  internalTrafficPolicy: Local
  ports:
  - port: 80
    targetPort: 80
    name: http
  selector:
    app: $NAME
---
apiVersion: v1
kind: Service
metadata:
  name: $NAME-external-v4
  namespace: ${NAMESPACE}
spec:
  type: NodePort
  externalIPs:
  - 169.254.0.90
  ports:
  - port: 80
    targetPort: 80
    name: http
  selector:
    app: $NAME
---
apiVersion: v1
kind: Service
metadata:
  name: $NAME-v6
  namespace: ${NAMESPACE}
spec:
  type: LoadBalancer
  ipFamilyPolicy: SingleStack
  ipFamilies:
  - IPv6
  ports:
  - port: 80
    targetPort: 80
    name: http
  selector:
    app: $NAME
EOF

