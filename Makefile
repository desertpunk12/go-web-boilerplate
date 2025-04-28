tg:
	templ generate --watch

tcss:
	tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css

server:
	air \
	build.cmd "go build -o main.exe cmd/hrapp/main.go" \
	--build.bin "main.exe" \
    --build.delay "100" \
    --build.exclude_dir "node_modules" \
    --build.include_ext "go" \
    --build.stop_on_error "false" \
    --misc.clean_on_exit true
