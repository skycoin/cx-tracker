module github.com/skycoin/cx-tracker

go 1.15

require (
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/SkycoinProject/cx v0.7.2-0.20201022102643-8d548a697fa8
	github.com/SkycoinProject/cx-chains v0.24.2-0.20200412040944-7696b1dfd81c
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/boltdb/bolt v1.3.1 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	go.etcd.io/bbolt v1.3.5
)

replace github.com/SkycoinProject/cx => ../cx
