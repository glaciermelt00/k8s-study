#!/bin/bash

# Ingress アクセスのための minikube tunnel を開始するスクリプト

echo "Starting minikube tunnel for Ingress access..."
echo "This will require sudo permissions for ports 80 and 443"
echo ""
echo "After the tunnel starts, you can access the API at:"
echo "  http://sm-api.local/health"
echo "  http://sm-api.local/metrics"
echo ""
echo "Press Ctrl+C to stop the tunnel"
echo ""

# minikube tunnel を実行（sudo が必要）
sudo minikube tunnel
