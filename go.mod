module github.com/akitasoftware/akita-libs

go 1.15

require (
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/gopacket v1.1.18
	github.com/google/martian/v3 v3.0.1
	github.com/google/uuid v1.2.0
	github.com/pkg/errors v0.9.1
)

replace (
	github.com/google/gopacket v1.1.18 => github.com/akitasoftware/gopacket v1.1.18-0.20201119235945-f584f5125293
	github.com/google/martian/v3 v3.0.1 => github.com/akitasoftware/martian/v3 v3.0.1-0.20210108221002-22c39e10ccd2
)
