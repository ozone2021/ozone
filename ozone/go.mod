module ozone-lib

go 1.14

require (
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.4.0
    ozone-daemon-lib v1.0.0 // indirect
)

replace ozone-daemon-lib => ../ozone-daemon-lib
