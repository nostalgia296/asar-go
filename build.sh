goVersion=$(go version | sed 's/go version //')
builtAt="$(date +'%F %T %z')"
ldflags="\
-w -s \
-X 'main.GoVersion=$goVersion' \
-X 'main.BuildAt=$builtAt' \
"
go build -ldflags="$ldflags"
