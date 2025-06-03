tg:
	templ generate --watch --proxy="http://localhost:3001" --cmd="go run ./cmd/hrapp-web/main.go"

be:
	go run ./cmd/hrapp-api/main.go

tcss:
	tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css -v --watch --minify


s:
	make -j3 tg tcss be
