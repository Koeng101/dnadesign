module github.com/koeng101/dnadesign/parts

go 1.22.5

require (
	github.com/koeng101/dnadesign/lib v0.0.0-00010101000000-000000000000
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/koeng101/dnadesign/lib => ../lib
