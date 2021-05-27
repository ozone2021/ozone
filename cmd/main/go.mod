module ozone

go 1.14

require (
	cli v1.0.0
	github.com/common-nighthawk/go-figure v0.0.0-20200609044655-c4b36f998cf2 // indirect
	github.com/jessevdk/go-flags v1.4.0 // indirect
	gopkg.in/src-d/go-git.v4 v4.13.1 // indirect
	ozone-daemon-lib v1.0.0 // indirect
	ozone-lib v1.0.0 // indirect
)

replace cli => ../cli

replace ozone-lib => ../../ozone

replace ozone-daemon-lib => ../../ozone-daemon-lib
