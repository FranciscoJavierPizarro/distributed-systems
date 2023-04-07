kubectl delete statefulset nodo
kubectl delete service raft
kubectl delete pod t
kubectl create -f raftv2.yaml

#kubectl get pod -o=custom-columns=NODE:.spec.nodeName,NAME:.metadata.name