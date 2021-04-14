module ozone-lib

go 1.14

require (
	github.com/joho/godotenv v1.3.0
	github.com/mitchellh/go-homedir v1.1.0
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.4.0
	ozone-daemon-lib v1.0.0
)

replace ozone-daemon-lib => ../ozone-daemon-lib
