#!/bin/bash

echo "🔄 refreshing vendor directory"
go mod vendor
go generate -mod=vendor ./pkg/...
git add --all
git commit -m "Vendor directory refresh"