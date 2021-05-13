module ozone-cli

go 1.14

require (
	github.com/flosch/pongo2/v4 v4.0.2 // indirect
	github.com/spf13/cobra v1.1.3
	ozone-daemon-lib v1.0.0
	ozone-lib v1.0.0
)

replace ozone-lib => ../../ozone

replace ozone-daemon-lib => ../../ozone-daemon-lib
