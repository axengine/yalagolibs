# etcd 权限配置
etcdctl --cert="./certs/client.pem" --key="./certs/client-key.pem" --cacert="./certs/ca.pem" --endpoints=https://192.168.1.51:2379,https://192.168.1.53:2379,https://192.168.1.105:2379 user add root
user add root
# 000000
user grant-role root root

添加角色
role add role-bridge
role grant-permission role-bridge readwrite /yala/bridge/ --prefix=true

添加用户
user add notary0 #000000
user add notary1
user add notary2

角色绑定
user grant-role notary0 role-bridge
user grant-role notary1 role-bridge
user grant-role notary2 role-bridge

启用用户名密码验证
auth enable