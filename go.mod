module pflow_objstorage

go 1.15

require (
	github.com/UCLabNU/proto_pflow v0.0.5
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/protobuf v1.4.3
	github.com/google/go-cmp v0.5.2 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/shirou/gopsutil v3.20.12+incompatible // indirect
	github.com/synerex/proto_pcounter v0.0.6
	github.com/synerex/proto_storage v0.2.0
	github.com/synerex/synerex_api v0.4.2
	github.com/synerex/synerex_proto v0.1.10
	github.com/synerex/synerex_sxutil v0.6.2
	golang.org/x/net v0.0.0-20201224014010-6772e930b67b // indirect
	golang.org/x/sys v0.0.0-20210108172913-0df2131ae363 // indirect
	golang.org/x/text v0.3.5 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/genproto v0.0.0-20210108203827-ffc7fda8c3d7 // indirect
	google.golang.org/grpc v1.34.0 // indirect
	gopkg.in/check.v1 v1.0.0-20200902074654-038fdea0a05b // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
)

replace github.com/synerex/proto_pcounter v0.0.6 => github.com/nagata-yoshiteru/proto_pcounter v0.0.9

replace github.com/synerex/synerex_proto v0.1.9 => github.com/nagata-yoshiteru/synerex_proto v0.1.12
