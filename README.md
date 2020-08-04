# ppw-operator

Operator for https://github.com/safronovD/python-pravega-writer

## Deploy
```bash
make generate
make manifests
make install
docker build . -t dxd360/ppw-operator:0.0.1
docker push dxd360/ppw-operator:0.0.1
make deploy IMG=dxd360/ppw-operator:0.0.1
kubectl apply -f config/samples/apps_v1alpha0_ppw.yaml
```